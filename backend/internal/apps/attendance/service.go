package attendance

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/config"
	"github.com/your-org/kintai/backend/internal/model"
	"github.com/your-org/kintai/backend/pkg/logger"
)

// NotificationSender は通知送信インターフェース（shared.NotificationService が実装）
type NotificationSender interface {
	Send(ctx context.Context, userID uuid.UUID, notifType model.NotificationType, title, message string) error
}

// エラー定義
var (
	ErrAlreadyClockedIn      = errors.New("既に出勤打刻済みです")
	ErrNotClockedIn          = errors.New("出勤打刻がありません")
	ErrAlreadyClockedOut     = errors.New("既に退勤打刻済みです")
	ErrLeaveNotFound         = errors.New("休暇申請が見つかりません")
	ErrLeaveAlreadyProcessed = errors.New("この休暇申請は既に処理済みです")
)

// Deps はサービスの依存関係
type Deps struct {
	Repos  *Repositories
	Config *config.Config
	Logger *logger.Logger
}

// Services は勤怠サービスを束ねる構造体
type Services struct {
	Attendance           AttendanceService
	Leave                LeaveService
	OvertimeRequest      OvertimeRequestService
	LeaveBalance         LeaveBalanceService
	AttendanceCorrection AttendanceCorrectionService
}

// NewServices は勤怠サービスを初期化する
func NewServices(deps Deps, notifier NotificationSender) *Services {
	return &Services{
		Attendance:           NewAttendanceService(deps),
		Leave:                NewLeaveService(deps, notifier),
		OvertimeRequest:      NewOvertimeRequestService(deps, notifier),
		LeaveBalance:         NewLeaveBalanceService(deps),
		AttendanceCorrection: NewAttendanceCorrectionService(deps, notifier),
	}
}

// ===== AttendanceService =====

type AttendanceService interface {
	ClockIn(ctx context.Context, userID uuid.UUID, req *model.ClockInRequest) (*model.Attendance, error)
	ClockOut(ctx context.Context, userID uuid.UUID, req *model.ClockOutRequest) (*model.Attendance, error)
	GetByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time, page, pageSize int) ([]model.Attendance, int64, error)
	GetSummary(ctx context.Context, userID uuid.UUID, start, end time.Time) (*model.AttendanceSummary, error)
	GetTodayStatus(ctx context.Context, userID uuid.UUID) (*model.Attendance, error)
}

type attendanceService struct {
	deps Deps
}

func NewAttendanceService(deps Deps) AttendanceService {
	return &attendanceService{deps: deps}
}

func (s *attendanceService) ClockIn(ctx context.Context, userID uuid.UUID, req *model.ClockInRequest) (*model.Attendance, error) {
	today := time.Now().Truncate(24 * time.Hour)

	existing, _ := s.deps.Repos.Attendance.FindByUserAndDate(ctx, userID, today)
	if existing != nil && existing.ClockIn != nil {
		return nil, ErrAlreadyClockedIn
	}

	now := time.Now()
	attendance := &model.Attendance{
		UserID:  userID,
		Date:    today,
		ClockIn: &now,
		Status:  model.AttendanceStatusPresent,
		Note:    req.Note,
	}

	if err := s.deps.Repos.Attendance.Create(ctx, attendance); err != nil {
		return nil, err
	}

	return attendance, nil
}

func (s *attendanceService) ClockOut(ctx context.Context, userID uuid.UUID, req *model.ClockOutRequest) (*model.Attendance, error) {
	today := time.Now().Truncate(24 * time.Hour)

	attendance, err := s.deps.Repos.Attendance.FindByUserAndDate(ctx, userID, today)
	if err != nil {
		return nil, ErrNotClockedIn
	}

	if attendance.ClockOut != nil {
		return nil, ErrAlreadyClockedOut
	}

	now := time.Now()
	attendance.ClockOut = &now
	if req.Note != "" {
		attendance.Note = req.Note
	}

	// 勤務時間を計算
	if attendance.ClockIn != nil {
		workDuration := now.Sub(*attendance.ClockIn)
		attendance.WorkMinutes = int(workDuration.Minutes()) - attendance.BreakMinutes

		// 8時間超過を残業として計算
		standardMinutes := 8 * 60
		if attendance.WorkMinutes > standardMinutes {
			attendance.OvertimeMinutes = attendance.WorkMinutes - standardMinutes
		}
	}

	if err := s.deps.Repos.Attendance.Update(ctx, attendance); err != nil {
		return nil, err
	}

	return attendance, nil
}

func (s *attendanceService) GetByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time, page, pageSize int) ([]model.Attendance, int64, error) {
	return s.deps.Repos.Attendance.FindByUserAndDateRange(ctx, userID, start, end, page, pageSize)
}

func (s *attendanceService) GetSummary(ctx context.Context, userID uuid.UUID, start, end time.Time) (*model.AttendanceSummary, error) {
	return s.deps.Repos.Attendance.GetSummary(ctx, userID, start, end)
}

func (s *attendanceService) GetTodayStatus(ctx context.Context, userID uuid.UUID) (*model.Attendance, error) {
	today := time.Now().Truncate(24 * time.Hour)
	return s.deps.Repos.Attendance.FindByUserAndDate(ctx, userID, today)
}

// ===== LeaveService =====

type LeaveService interface {
	Create(ctx context.Context, userID uuid.UUID, req *model.LeaveRequestCreate) (*model.LeaveRequest, error)
	Approve(ctx context.Context, leaveID uuid.UUID, approverID uuid.UUID, req *model.LeaveRequestApproval) (*model.LeaveRequest, error)
	GetByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.LeaveRequest, int64, error)
	GetPending(ctx context.Context, page, pageSize int) ([]model.LeaveRequest, int64, error)
}

type leaveService struct {
	deps     Deps
	notifier NotificationSender
}

func NewLeaveService(deps Deps, notifier NotificationSender) LeaveService {
	return &leaveService{deps: deps, notifier: notifier}
}

func (s *leaveService) Create(ctx context.Context, userID uuid.UUID, req *model.LeaveRequestCreate) (*model.LeaveRequest, error) {
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, errors.New("開始日の形式が不正です")
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, errors.New("終了日の形式が不正です")
	}

	leave := &model.LeaveRequest{
		UserID:    userID,
		LeaveType: req.LeaveType,
		StartDate: startDate,
		EndDate:   endDate,
		Reason:    req.Reason,
		Status:    model.ApprovalStatusPending,
	}

	if err := s.deps.Repos.LeaveRequest.Create(ctx, leave); err != nil {
		return nil, err
	}

	return leave, nil
}

func (s *leaveService) Approve(ctx context.Context, leaveID uuid.UUID, approverID uuid.UUID, req *model.LeaveRequestApproval) (*model.LeaveRequest, error) {
	leave, err := s.deps.Repos.LeaveRequest.FindByID(ctx, leaveID)
	if err != nil {
		return nil, ErrLeaveNotFound
	}

	if leave.Status != model.ApprovalStatusPending {
		return nil, ErrLeaveAlreadyProcessed
	}

	now := time.Now()
	leave.Status = req.Status
	leave.ApprovedBy = &approverID
	leave.ApprovedAt = &now
	if req.Status == model.ApprovalStatusRejected {
		leave.RejectedReason = req.RejectedReason
	}

	if err := s.deps.Repos.LeaveRequest.Update(ctx, leave); err != nil {
		return nil, err
	}

	// 申請者に通知を送信
	if req.Status == model.ApprovalStatusApproved {
		_ = s.notifier.Send(ctx, leave.UserID, model.NotificationTypeLeaveApproved,
			"休暇申請が承認されました",
			fmt.Sprintf("あなたの休暇申請（%s〜%s）が承認されました。", leave.StartDate.Format("2006-01-02"), leave.EndDate.Format("2006-01-02")))
	} else if req.Status == model.ApprovalStatusRejected {
		msg := fmt.Sprintf("あなたの休暇申請（%s〜%s）が却下されました。", leave.StartDate.Format("2006-01-02"), leave.EndDate.Format("2006-01-02"))
		if req.RejectedReason != "" {
			msg += " 理由: " + req.RejectedReason
		}
		_ = s.notifier.Send(ctx, leave.UserID, model.NotificationTypeLeaveRejected, "休暇申請が却下されました", msg)
	}

	return leave, nil
}

func (s *leaveService) GetByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.LeaveRequest, int64, error) {
	return s.deps.Repos.LeaveRequest.FindByUserID(ctx, userID, page, pageSize)
}

func (s *leaveService) GetPending(ctx context.Context, page, pageSize int) ([]model.LeaveRequest, int64, error) {
	return s.deps.Repos.LeaveRequest.FindPending(ctx, page, pageSize)
}

// ===== OvertimeRequestService =====

type OvertimeRequestService interface {
	Create(ctx context.Context, userID uuid.UUID, req *model.OvertimeRequestCreate) (*model.OvertimeRequest, error)
	Approve(ctx context.Context, id uuid.UUID, approverID uuid.UUID, req *model.OvertimeRequestApproval) (*model.OvertimeRequest, error)
	GetByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.OvertimeRequest, int64, error)
	GetPending(ctx context.Context, page, pageSize int) ([]model.OvertimeRequest, int64, error)
	GetOvertimeAlerts(ctx context.Context) ([]model.OvertimeAlert, error)
}

type overtimeRequestService struct {
	deps     Deps
	notifier NotificationSender
}

func NewOvertimeRequestService(deps Deps, notifier NotificationSender) OvertimeRequestService {
	return &overtimeRequestService{deps: deps, notifier: notifier}
}

func (s *overtimeRequestService) Create(ctx context.Context, userID uuid.UUID, req *model.OvertimeRequestCreate) (*model.OvertimeRequest, error) {
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, errors.New("日付の形式が不正です")
	}
	overtime := &model.OvertimeRequest{
		UserID: userID, Date: date, PlannedMinutes: req.PlannedMinutes,
		Reason: req.Reason, Status: model.OvertimeStatusPending,
	}
	if err := s.deps.Repos.OvertimeRequest.Create(ctx, overtime); err != nil {
		return nil, err
	}
	return overtime, nil
}

func (s *overtimeRequestService) Approve(ctx context.Context, id uuid.UUID, approverID uuid.UUID, req *model.OvertimeRequestApproval) (*model.OvertimeRequest, error) {
	overtime, err := s.deps.Repos.OvertimeRequest.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("残業申請が見つかりません")
	}
	if overtime.Status != model.OvertimeStatusPending {
		return nil, errors.New("この残業申請は既に処理済みです")
	}
	now := time.Now()
	overtime.Status = req.Status
	overtime.ApprovedBy = &approverID
	overtime.ApprovedAt = &now
	if req.Status == model.OvertimeStatusRejected {
		overtime.RejectedReason = req.RejectedReason
	}
	if err := s.deps.Repos.OvertimeRequest.Update(ctx, overtime); err != nil {
		return nil, err
	}
	// 通知送信
	notifType := model.NotificationTypeLeaveApproved
	title := "残業申請が承認されました"
	if req.Status == model.OvertimeStatusRejected {
		notifType = model.NotificationTypeLeaveRejected
		title = "残業申請が却下されました"
	}
	_ = s.notifier.Send(ctx, overtime.UserID, notifType, title,
		fmt.Sprintf("%s の残業申請が%sされました", overtime.Date.Format("2006-01-02"),
			map[model.OvertimeRequestStatus]string{model.OvertimeStatusApproved: "承認", model.OvertimeStatusRejected: "却下"}[req.Status]))
	return overtime, nil
}

func (s *overtimeRequestService) GetByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.OvertimeRequest, int64, error) {
	return s.deps.Repos.OvertimeRequest.FindByUserID(ctx, userID, page, pageSize)
}

func (s *overtimeRequestService) GetPending(ctx context.Context, page, pageSize int) ([]model.OvertimeRequest, int64, error) {
	return s.deps.Repos.OvertimeRequest.FindPending(ctx, page, pageSize)
}

func (s *overtimeRequestService) GetOvertimeAlerts(ctx context.Context) ([]model.OvertimeAlert, error) {
	now := time.Now()
	users, _, err := s.deps.Repos.User.FindAll(ctx, 1, 10000)
	if err != nil {
		return nil, err
	}

	alerts := make([]model.OvertimeAlert, 0)
	const monthlyLimit = 35.0
	const yearlyLimit = 360.0

	for _, u := range users {
		monthly, err := s.deps.Repos.OvertimeRequest.GetUserMonthlyOvertime(ctx, u.ID, now.Year(), int(now.Month()))
		if err != nil {
			return nil, err
		}
		yearly, err := s.deps.Repos.OvertimeRequest.GetUserYearlyOvertime(ctx, u.ID, now.Year())
		if err != nil {
			return nil, err
		}

		monthlyHours := float64(monthly) / 60.0
		yearlyHours := float64(yearly) / 60.0
		isMonthlyExceeded := monthlyHours > monthlyLimit
		isYearlyExceeded := yearlyHours > yearlyLimit
		if !isMonthlyExceeded && !isYearlyExceeded {
			continue
		}

		alerts = append(alerts, model.OvertimeAlert{
			UserID:               u.ID,
			UserName:             u.LastName + " " + u.FirstName,
			MonthlyOvertimeHours: monthlyHours,
			YearlyOvertimeHours:  yearlyHours,
			MonthlyLimitHours:    monthlyLimit,
			YearlyLimitHours:     yearlyLimit,
			IsMonthlyExceeded:    isMonthlyExceeded,
			IsYearlyExceeded:     isYearlyExceeded,
		})
	}

	return alerts, nil
}

// ===== LeaveBalanceService =====

type LeaveBalanceService interface {
	GetByUser(ctx context.Context, userID uuid.UUID, fiscalYear int) ([]model.LeaveBalanceResponse, error)
	SetBalance(ctx context.Context, userID uuid.UUID, fiscalYear int, leaveType model.LeaveType, req *model.LeaveBalanceUpdate) error
	DeductBalance(ctx context.Context, userID uuid.UUID, leaveType model.LeaveType, days float64) error
	InitializeForUser(ctx context.Context, userID uuid.UUID, fiscalYear int) error
}

type leaveBalanceService struct{ deps Deps }

func NewLeaveBalanceService(deps Deps) LeaveBalanceService {
	return &leaveBalanceService{deps: deps}
}

func (s *leaveBalanceService) GetByUser(ctx context.Context, userID uuid.UUID, fiscalYear int) ([]model.LeaveBalanceResponse, error) {
	balances, err := s.deps.Repos.LeaveBalance.FindByUserAndYear(ctx, userID, fiscalYear)
	if err != nil {
		return nil, err
	}
	var responses []model.LeaveBalanceResponse
	for _, b := range balances {
		responses = append(responses, model.LeaveBalanceResponse{
			LeaveType: b.LeaveType, TotalDays: b.TotalDays, UsedDays: b.UsedDays,
			RemainingDays: b.TotalDays + b.CarriedOver - b.UsedDays,
			CarriedOver:   b.CarriedOver, FiscalYear: b.FiscalYear,
		})
	}
	return responses, nil
}

func (s *leaveBalanceService) SetBalance(ctx context.Context, userID uuid.UUID, fiscalYear int, leaveType model.LeaveType, req *model.LeaveBalanceUpdate) error {
	balance, err := s.deps.Repos.LeaveBalance.FindByUserYearAndType(ctx, userID, fiscalYear, leaveType)
	if err != nil {
		balance = &model.LeaveBalance{UserID: userID, FiscalYear: fiscalYear, LeaveType: leaveType}
	}
	if req.TotalDays != nil {
		balance.TotalDays = *req.TotalDays
	}
	if req.CarriedOver != nil {
		balance.CarriedOver = *req.CarriedOver
	}
	return s.deps.Repos.LeaveBalance.Upsert(ctx, balance)
}

func (s *leaveBalanceService) DeductBalance(ctx context.Context, userID uuid.UUID, leaveType model.LeaveType, days float64) error {
	fiscalYear := time.Now().Year()
	balance, err := s.deps.Repos.LeaveBalance.FindByUserYearAndType(ctx, userID, fiscalYear, leaveType)
	if err != nil {
		return errors.New("有給残日数が設定されていません")
	}
	remaining := balance.TotalDays + balance.CarriedOver - balance.UsedDays
	if remaining < days {
		return fmt.Errorf("有給残日数が不足しています（残り: %.1f日）", remaining)
	}
	balance.UsedDays += days
	return s.deps.Repos.LeaveBalance.Update(ctx, balance)
}

func (s *leaveBalanceService) InitializeForUser(ctx context.Context, userID uuid.UUID, fiscalYear int) error {
	defaultDays := map[model.LeaveType]float64{
		model.LeaveTypePaid: 10, model.LeaveTypeSick: 5, model.LeaveTypeSpecial: 3,
	}
	for lt, days := range defaultDays {
		balance := &model.LeaveBalance{
			UserID: userID, FiscalYear: fiscalYear, LeaveType: lt,
			TotalDays: days, UsedDays: 0, CarriedOver: 0,
		}
		if err := s.deps.Repos.LeaveBalance.Upsert(ctx, balance); err != nil {
			return err
		}
	}
	return nil
}

// ===== AttendanceCorrectionService =====

type AttendanceCorrectionService interface {
	Create(ctx context.Context, userID uuid.UUID, req *model.AttendanceCorrectionCreate) (*model.AttendanceCorrection, error)
	Approve(ctx context.Context, id uuid.UUID, approverID uuid.UUID, req *model.AttendanceCorrectionApproval) (*model.AttendanceCorrection, error)
	GetByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.AttendanceCorrection, int64, error)
	GetPending(ctx context.Context, page, pageSize int) ([]model.AttendanceCorrection, int64, error)
}

type attendanceCorrectionService struct {
	deps     Deps
	notifier NotificationSender
}

func NewAttendanceCorrectionService(deps Deps, notifier NotificationSender) AttendanceCorrectionService {
	return &attendanceCorrectionService{deps: deps, notifier: notifier}
}

func (s *attendanceCorrectionService) Create(ctx context.Context, userID uuid.UUID, req *model.AttendanceCorrectionCreate) (*model.AttendanceCorrection, error) {
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, errors.New("日付の形式が不正です")
	}
	correction := &model.AttendanceCorrection{
		UserID: userID, Date: date, Reason: req.Reason,
		Status: model.CorrectionStatusPending,
	}
	// 既存の出勤データを取得
	existing, _ := s.deps.Repos.Attendance.FindByUserAndDate(ctx, userID, date)
	if existing != nil {
		correction.AttendanceID = &existing.ID
		correction.OriginalClockIn = existing.ClockIn
		correction.OriginalClockOut = existing.ClockOut
	}
	if req.CorrectedClockIn != nil {
		t, err := time.Parse("2006-01-02T15:04:05", *req.CorrectedClockIn)
		if err != nil {
			t, err = time.Parse("15:04", *req.CorrectedClockIn)
			if err == nil {
				t = time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, time.Local)
			}
		}
		if err == nil {
			correction.CorrectedClockIn = &t
		}
	}
	if req.CorrectedClockOut != nil {
		t, err := time.Parse("2006-01-02T15:04:05", *req.CorrectedClockOut)
		if err != nil {
			t, err = time.Parse("15:04", *req.CorrectedClockOut)
			if err == nil {
				t = time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, time.Local)
			}
		}
		if err == nil {
			correction.CorrectedClockOut = &t
		}
	}
	if err := s.deps.Repos.AttendanceCorrection.Create(ctx, correction); err != nil {
		return nil, err
	}
	return correction, nil
}

func (s *attendanceCorrectionService) Approve(ctx context.Context, id uuid.UUID, approverID uuid.UUID, req *model.AttendanceCorrectionApproval) (*model.AttendanceCorrection, error) {
	correction, err := s.deps.Repos.AttendanceCorrection.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("修正申請が見つかりません")
	}
	if correction.Status != model.CorrectionStatusPending {
		return nil, errors.New("この修正申請は既に処理済みです")
	}
	now := time.Now()
	correction.Status = req.Status
	correction.ApprovedBy = &approverID
	correction.ApprovedAt = &now
	if req.Status == model.CorrectionStatusRejected {
		correction.RejectedReason = req.RejectedReason
	}
	// 承認時：勤怠データを修正
	if req.Status == model.CorrectionStatusApproved {
		if correction.AttendanceID != nil {
			att, _ := s.deps.Repos.Attendance.FindByID(ctx, *correction.AttendanceID)
			if att != nil {
				if correction.CorrectedClockIn != nil {
					att.ClockIn = correction.CorrectedClockIn
				}
				if correction.CorrectedClockOut != nil {
					att.ClockOut = correction.CorrectedClockOut
				}
				if att.ClockIn != nil && att.ClockOut != nil {
					workDuration := att.ClockOut.Sub(*att.ClockIn)
					att.WorkMinutes = int(workDuration.Minutes()) - att.BreakMinutes
					if att.WorkMinutes > 480 {
						att.OvertimeMinutes = att.WorkMinutes - 480
					} else {
						att.OvertimeMinutes = 0
					}
				}
				_ = s.deps.Repos.Attendance.Update(ctx, att)
			}
		} else {
			// 新規作成
			att := &model.Attendance{
				UserID: correction.UserID, Date: correction.Date,
				ClockIn: correction.CorrectedClockIn, ClockOut: correction.CorrectedClockOut,
				Status: model.AttendanceStatusPresent,
			}
			if att.ClockIn != nil && att.ClockOut != nil {
				workDuration := att.ClockOut.Sub(*att.ClockIn)
				att.WorkMinutes = int(workDuration.Minutes())
				if att.WorkMinutes > 480 {
					att.OvertimeMinutes = att.WorkMinutes - 480
				}
			}
			_ = s.deps.Repos.Attendance.Create(ctx, att)
		}
	}
	if err := s.deps.Repos.AttendanceCorrection.Update(ctx, correction); err != nil {
		return nil, err
	}
	// 通知送信
	title := "勤怠修正申請が承認されました"
	if req.Status == model.CorrectionStatusRejected {
		title = "勤怠修正申請が却下されました"
	}
	_ = s.notifier.Send(ctx, correction.UserID, model.NotificationTypeCorrectionResult, title,
		fmt.Sprintf("%s の勤怠修正申請が処理されました", correction.Date.Format("2006-01-02")))
	return correction, nil
}

func (s *attendanceCorrectionService) GetByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]model.AttendanceCorrection, int64, error) {
	return s.deps.Repos.AttendanceCorrection.FindByUserID(ctx, userID, page, pageSize)
}

func (s *attendanceCorrectionService) GetPending(ctx context.Context, page, pageSize int) ([]model.AttendanceCorrection, int64, error) {
	return s.deps.Repos.AttendanceCorrection.FindPending(ctx, page, pageSize)
}
