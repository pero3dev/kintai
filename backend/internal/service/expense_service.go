package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/model"
)

// ===== ExpenseService =====

type ExpenseService interface {
	Create(ctx context.Context, userID uuid.UUID, req *model.ExpenseCreateRequest) (*model.Expense, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Expense, error)
	GetList(ctx context.Context, userID uuid.UUID, page, pageSize int, status, category string) ([]model.Expense, int64, error)
	Update(ctx context.Context, id uuid.UUID, userID uuid.UUID, req *model.ExpenseUpdateRequest) (*model.Expense, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetPending(ctx context.Context, page, pageSize int) ([]model.Expense, int64, error)
	Approve(ctx context.Context, id uuid.UUID, approverID uuid.UUID, req *model.ExpenseApproveRequest) error
	AdvancedApprove(ctx context.Context, id uuid.UUID, approverID uuid.UUID, req *model.ExpenseAdvancedApproveRequest) error
	GetStats(ctx context.Context, userID uuid.UUID) (*model.ExpenseStatsResponse, error)
	GetReport(ctx context.Context, startDate, endDate string) (*model.ExpenseReportResponse, error)
	GetMonthlyTrend(ctx context.Context, year string) ([]model.MonthlyTrendItem, error)
	ExportCSV(ctx context.Context, startDate, endDate string) ([]byte, error)
	ExportPDF(ctx context.Context, startDate, endDate string) ([]byte, error)
}

type expenseService struct {
	deps Deps
}

func NewExpenseService(deps Deps) ExpenseService {
	return &expenseService{deps: deps}
}

func (s *expenseService) Create(ctx context.Context, userID uuid.UUID, req *model.ExpenseCreateRequest) (*model.Expense, error) {
	expense := &model.Expense{
		UserID: userID,
		Title:  req.Title,
		Status: model.ExpenseStatus(req.Status),
		Notes:  req.Notes,
	}
	if expense.Status == "" {
		expense.Status = model.ExpenseStatusDraft
	}

	var totalAmount float64
	for _, item := range req.Items {
		expDate, _ := time.Parse("2006-01-02", item.ExpenseDate)
		expense.Items = append(expense.Items, model.ExpenseItem{
			ExpenseDate: expDate,
			Category:    model.ExpenseCategory(item.Category),
			Description: item.Description,
			Amount:      item.Amount,
			ReceiptURL:  item.ReceiptURL,
		})
		totalAmount += item.Amount
	}
	expense.TotalAmount = totalAmount

	if err := s.deps.Repos.Expense.Create(ctx, expense); err != nil {
		return nil, err
	}

	// 履歴記録
	user, _ := s.deps.Repos.User.FindByID(ctx, userID)
	userName := ""
	if user != nil {
		userName = user.LastName + " " + user.FirstName
	}
	s.deps.Repos.ExpenseHistory.Create(ctx, &model.ExpenseHistory{
		ExpenseID: expense.ID,
		UserID:    userID,
		Action:    "作成",
		NewValue:  string(expense.Status),
		ChangedBy: userName,
	})

	return expense, nil
}

func (s *expenseService) GetByID(ctx context.Context, id uuid.UUID) (*model.Expense, error) {
	return s.deps.Repos.Expense.FindByID(ctx, id)
}

func (s *expenseService) GetList(ctx context.Context, userID uuid.UUID, page, pageSize int, status, category string) ([]model.Expense, int64, error) {
	return s.deps.Repos.Expense.FindByUserID(ctx, userID, page, pageSize, status, category)
}

func (s *expenseService) Update(ctx context.Context, id uuid.UUID, userID uuid.UUID, req *model.ExpenseUpdateRequest) (*model.Expense, error) {
	expense, err := s.deps.Repos.Expense.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldStatus := string(expense.Status)

	if req.Title != nil {
		expense.Title = *req.Title
	}
	if req.Status != nil {
		expense.Status = model.ExpenseStatus(*req.Status)
	}
	if req.Notes != nil {
		expense.Notes = *req.Notes
	}

	if req.Items != nil {
		// 明細を再作成
		s.deps.Repos.ExpenseItem.DeleteByExpenseID(ctx, id)
		var items []model.ExpenseItem
		var totalAmount float64
		for _, item := range req.Items {
			expDate, _ := time.Parse("2006-01-02", item.ExpenseDate)
			items = append(items, model.ExpenseItem{
				ExpenseID:   id,
				ExpenseDate: expDate,
				Category:    model.ExpenseCategory(item.Category),
				Description: item.Description,
				Amount:      item.Amount,
				ReceiptURL:  item.ReceiptURL,
			})
			totalAmount += item.Amount
		}
		s.deps.Repos.ExpenseItem.CreateBatch(ctx, items)
		expense.TotalAmount = totalAmount
	}

	if err := s.deps.Repos.Expense.Update(ctx, expense); err != nil {
		return nil, err
	}

	// 履歴記録
	user, _ := s.deps.Repos.User.FindByID(ctx, userID)
	userName := ""
	if user != nil {
		userName = user.LastName + " " + user.FirstName
	}
	s.deps.Repos.ExpenseHistory.Create(ctx, &model.ExpenseHistory{
		ExpenseID: id,
		UserID:    userID,
		Action:    "更新",
		OldValue:  oldStatus,
		NewValue:  string(expense.Status),
		ChangedBy: userName,
	})

	return expense, nil
}

func (s *expenseService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.deps.Repos.Expense.Delete(ctx, id)
}

func (s *expenseService) GetPending(ctx context.Context, page, pageSize int) ([]model.Expense, int64, error) {
	return s.deps.Repos.Expense.FindPending(ctx, page, pageSize)
}

func (s *expenseService) Approve(ctx context.Context, id uuid.UUID, approverID uuid.UUID, req *model.ExpenseApproveRequest) error {
	expense, err := s.deps.Repos.Expense.FindByID(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()
	expense.Status = model.ExpenseStatus(req.Status)
	expense.ApprovedBy = &approverID
	expense.ApprovedAt = &now
	if req.RejectedReason != "" {
		expense.RejectedReason = req.RejectedReason
	}

	if err := s.deps.Repos.Expense.Update(ctx, expense); err != nil {
		return err
	}

	// 通知作成
	notifType := "approved"
	message := fmt.Sprintf("経費申請「%s」が承認されました", expense.Title)
	if expense.Status == model.ExpenseStatusRejected {
		notifType = "rejected"
		message = fmt.Sprintf("経費申請「%s」が却下されました", expense.Title)
	}
	s.deps.Repos.ExpenseNotification.Create(ctx, &model.ExpenseNotification{
		UserID:    expense.UserID,
		ExpenseID: &expense.ID,
		Type:      notifType,
		Message:   message,
	})

	// 履歴記録
	approver, _ := s.deps.Repos.User.FindByID(ctx, approverID)
	approverName := ""
	if approver != nil {
		approverName = approver.LastName + " " + approver.FirstName
	}
	s.deps.Repos.ExpenseHistory.Create(ctx, &model.ExpenseHistory{
		ExpenseID: id,
		UserID:    approverID,
		Action:    "承認処理",
		NewValue:  req.Status,
		ChangedBy: approverName,
	})

	return nil
}

func (s *expenseService) AdvancedApprove(ctx context.Context, id uuid.UUID, approverID uuid.UUID, req *model.ExpenseAdvancedApproveRequest) error {
	expense, err := s.deps.Repos.Expense.FindByID(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()
	switch req.Action {
	case "approve":
		if req.Step == 1 {
			expense.Status = model.ExpenseStatusStep1Approved
		} else {
			expense.Status = model.ExpenseStatusApproved
			expense.ApprovedBy = &approverID
			expense.ApprovedAt = &now
		}
	case "reject":
		expense.Status = model.ExpenseStatusRejected
		expense.RejectedReason = req.Reason
	case "return":
		expense.Status = model.ExpenseStatusReturned
	}

	if err := s.deps.Repos.Expense.Update(ctx, expense); err != nil {
		return err
	}

	// 履歴記録
	approver, _ := s.deps.Repos.User.FindByID(ctx, approverID)
	approverName := ""
	if approver != nil {
		approverName = approver.LastName + " " + approver.FirstName
	}
	s.deps.Repos.ExpenseHistory.Create(ctx, &model.ExpenseHistory{
		ExpenseID: id,
		UserID:    approverID,
		Action:    "高度承認: " + req.Action,
		NewValue:  string(expense.Status),
		ChangedBy: approverName,
	})

	return nil
}

func (s *expenseService) GetStats(ctx context.Context, userID uuid.UUID) (*model.ExpenseStatsResponse, error) {
	return s.deps.Repos.Expense.GetStats(ctx, userID)
}

func (s *expenseService) GetReport(ctx context.Context, startDateStr, endDateStr string) (*model.ExpenseReportResponse, error) {
	startDate, _ := time.Parse("2006-01-02", startDateStr)
	endDate, _ := time.Parse("2006-01-02", endDateStr)
	endDate = endDate.Add(24*time.Hour - time.Second)
	return s.deps.Repos.Expense.GetReport(ctx, startDate, endDate)
}

func (s *expenseService) GetMonthlyTrend(ctx context.Context, yearStr string) ([]model.MonthlyTrendItem, error) {
	year, _ := strconv.Atoi(yearStr)
	if year == 0 {
		year = time.Now().Year()
	}
	return s.deps.Repos.Expense.GetMonthlyTrend(ctx, year)
}

func (s *expenseService) ExportCSV(ctx context.Context, startDateStr, endDateStr string) ([]byte, error) {
	startDate, _ := time.Parse("2006-01-02", startDateStr)
	endDate, _ := time.Parse("2006-01-02", endDateStr)
	endDate = endDate.Add(24*time.Hour - time.Second)

	expenses, _, err := s.deps.Repos.Expense.FindAll(ctx, 1, 10000, "", "")
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	// BOM for Excel
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	writer.Write([]string{"ID", "タイトル", "申請者", "ステータス", "合計金額", "作成日"})
	for _, exp := range expenses {
		if exp.CreatedAt.Before(startDate) || exp.CreatedAt.After(endDate) {
			continue
		}
		userName := ""
		if exp.User != nil {
			userName = exp.User.LastName + " " + exp.User.FirstName
		}
		writer.Write([]string{
			exp.ID.String(),
			exp.Title,
			userName,
			string(exp.Status),
			fmt.Sprintf("%.0f", exp.TotalAmount),
			exp.CreatedAt.Format("2006-01-02"),
		})
	}
	writer.Flush()
	return buf.Bytes(), nil
}

func (s *expenseService) ExportPDF(ctx context.Context, startDateStr, endDateStr string) ([]byte, error) {
	// PDFエクスポートは簡易的にCSVと同じデータをテキストで返す
	return s.ExportCSV(ctx, startDateStr, endDateStr)
}

// ===== ExpenseCommentService =====

type ExpenseCommentService interface {
	GetComments(ctx context.Context, expenseID uuid.UUID) ([]model.ExpenseCommentResponse, error)
	AddComment(ctx context.Context, expenseID, userID uuid.UUID, req *model.ExpenseCommentRequest) (*model.ExpenseCommentResponse, error)
}

type expenseCommentService struct {
	deps Deps
}

func NewExpenseCommentService(deps Deps) ExpenseCommentService {
	return &expenseCommentService{deps: deps}
}

func (s *expenseCommentService) GetComments(ctx context.Context, expenseID uuid.UUID) ([]model.ExpenseCommentResponse, error) {
	comments, err := s.deps.Repos.ExpenseComment.FindByExpenseID(ctx, expenseID)
	if err != nil {
		return nil, err
	}

	var resp []model.ExpenseCommentResponse
	for _, c := range comments {
		userName := ""
		if c.User != nil {
			userName = c.User.LastName + " " + c.User.FirstName
		}
		resp = append(resp, model.ExpenseCommentResponse{
			ID:        c.ID,
			ExpenseID: c.ExpenseID,
			UserID:    c.UserID,
			UserName:  userName,
			Content:   c.Content,
			CreatedAt: c.CreatedAt,
		})
	}
	if resp == nil {
		resp = []model.ExpenseCommentResponse{}
	}
	return resp, nil
}

func (s *expenseCommentService) AddComment(ctx context.Context, expenseID, userID uuid.UUID, req *model.ExpenseCommentRequest) (*model.ExpenseCommentResponse, error) {
	comment := &model.ExpenseComment{
		ExpenseID: expenseID,
		UserID:    userID,
		Content:   req.Content,
	}
	if err := s.deps.Repos.ExpenseComment.Create(ctx, comment); err != nil {
		return nil, err
	}

	user, _ := s.deps.Repos.User.FindByID(ctx, userID)
	userName := ""
	if user != nil {
		userName = user.LastName + " " + user.FirstName
	}

	// 通知
	expense, _ := s.deps.Repos.Expense.FindByID(ctx, expenseID)
	if expense != nil && expense.UserID != userID {
		s.deps.Repos.ExpenseNotification.Create(ctx, &model.ExpenseNotification{
			UserID:    expense.UserID,
			ExpenseID: &expenseID,
			Type:      "comment",
			Message:   fmt.Sprintf("%sが経費申請にコメントしました", userName),
		})
	}

	return &model.ExpenseCommentResponse{
		ID:        comment.ID,
		ExpenseID: comment.ExpenseID,
		UserID:    comment.UserID,
		UserName:  userName,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
	}, nil
}

// ===== ExpenseHistoryService =====

type ExpenseHistoryService interface {
	GetHistory(ctx context.Context, expenseID uuid.UUID) ([]model.ExpenseHistoryResponse, error)
}

type expenseHistoryService struct {
	deps Deps
}

func NewExpenseHistoryService(deps Deps) ExpenseHistoryService {
	return &expenseHistoryService{deps: deps}
}

func (s *expenseHistoryService) GetHistory(ctx context.Context, expenseID uuid.UUID) ([]model.ExpenseHistoryResponse, error) {
	histories, err := s.deps.Repos.ExpenseHistory.FindByExpenseID(ctx, expenseID)
	if err != nil {
		return nil, err
	}

	var resp []model.ExpenseHistoryResponse
	for _, h := range histories {
		resp = append(resp, model.ExpenseHistoryResponse{
			ID:        h.ID,
			ExpenseID: h.ExpenseID,
			Action:    h.Action,
			ChangedBy: h.ChangedBy,
			OldValue:  h.OldValue,
			NewValue:  h.NewValue,
			CreatedAt: h.CreatedAt,
		})
	}
	if resp == nil {
		resp = []model.ExpenseHistoryResponse{}
	}
	return resp, nil
}

// ===== ExpenseReceiptService =====

type ExpenseReceiptService interface {
	Upload(ctx context.Context, filename string, data []byte) (string, error)
}

type expenseReceiptService struct {
	deps Deps
}

func NewExpenseReceiptService(deps Deps) ExpenseReceiptService {
	return &expenseReceiptService{deps: deps}
}

func (s *expenseReceiptService) Upload(ctx context.Context, filename string, data []byte) (string, error) {
	// 簡易実装: ファイル名ベースのURLを返す（実運用ではS3等に保存）
	url := fmt.Sprintf("/uploads/receipts/%s_%s", uuid.New().String()[:8], filename)
	return url, nil
}

// ===== ExpenseTemplateService =====

type ExpenseTemplateService interface {
	GetTemplates(ctx context.Context, userID uuid.UUID) ([]model.ExpenseTemplate, error)
	Create(ctx context.Context, userID uuid.UUID, req *model.ExpenseTemplateRequest) (*model.ExpenseTemplate, error)
	Update(ctx context.Context, id uuid.UUID, req *model.ExpenseTemplateRequest) (*model.ExpenseTemplate, error)
	Delete(ctx context.Context, id uuid.UUID) error
	UseTemplate(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*model.Expense, error)
}

type expenseTemplateService struct {
	deps Deps
}

func NewExpenseTemplateService(deps Deps) ExpenseTemplateService {
	return &expenseTemplateService{deps: deps}
}

func (s *expenseTemplateService) GetTemplates(ctx context.Context, userID uuid.UUID) ([]model.ExpenseTemplate, error) {
	return s.deps.Repos.ExpenseTemplate.FindAll(ctx, userID)
}

func (s *expenseTemplateService) Create(ctx context.Context, userID uuid.UUID, req *model.ExpenseTemplateRequest) (*model.ExpenseTemplate, error) {
	tmpl := &model.ExpenseTemplate{
		UserID:       userID,
		Name:         req.Name,
		Title:        req.Title,
		Category:     model.ExpenseCategory(req.Category),
		Description:  req.Description,
		Amount:       req.Amount,
		IsRecurring:  req.IsRecurring,
		RecurringDay: req.RecurringDay,
	}
	if err := s.deps.Repos.ExpenseTemplate.Create(ctx, tmpl); err != nil {
		return nil, err
	}
	return tmpl, nil
}

func (s *expenseTemplateService) Update(ctx context.Context, id uuid.UUID, req *model.ExpenseTemplateRequest) (*model.ExpenseTemplate, error) {
	tmpl, err := s.deps.Repos.ExpenseTemplate.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	tmpl.Name = req.Name
	tmpl.Title = req.Title
	tmpl.Category = model.ExpenseCategory(req.Category)
	tmpl.Description = req.Description
	tmpl.Amount = req.Amount
	tmpl.IsRecurring = req.IsRecurring
	tmpl.RecurringDay = req.RecurringDay

	if err := s.deps.Repos.ExpenseTemplate.Update(ctx, tmpl); err != nil {
		return nil, err
	}
	return tmpl, nil
}

func (s *expenseTemplateService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.deps.Repos.ExpenseTemplate.Delete(ctx, id)
}

func (s *expenseTemplateService) UseTemplate(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*model.Expense, error) {
	tmpl, err := s.deps.Repos.ExpenseTemplate.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	expense := &model.Expense{
		UserID:      userID,
		Title:       tmpl.Title,
		Status:      model.ExpenseStatusDraft,
		TotalAmount: tmpl.Amount,
		Items: []model.ExpenseItem{
			{
				ExpenseDate: time.Now(),
				Category:    tmpl.Category,
				Description: tmpl.Description,
				Amount:      tmpl.Amount,
			},
		},
	}
	if err := s.deps.Repos.Expense.Create(ctx, expense); err != nil {
		return nil, err
	}
	return expense, nil
}

// ===== ExpensePolicyService =====

type ExpensePolicyService interface {
	GetPolicies(ctx context.Context) ([]model.ExpensePolicy, error)
	Create(ctx context.Context, req *model.ExpensePolicyRequest) (*model.ExpensePolicy, error)
	Update(ctx context.Context, id uuid.UUID, req *model.ExpensePolicyRequest) (*model.ExpensePolicy, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetBudgets(ctx context.Context) ([]model.ExpenseBudget, error)
	GetPolicyViolations(ctx context.Context, userID uuid.UUID) ([]model.ExpensePolicyViolation, error)
}

type expensePolicyService struct {
	deps Deps
}

func NewExpensePolicyService(deps Deps) ExpensePolicyService {
	return &expensePolicyService{deps: deps}
}

func (s *expensePolicyService) GetPolicies(ctx context.Context) ([]model.ExpensePolicy, error) {
	return s.deps.Repos.ExpensePolicy.FindAll(ctx)
}

func (s *expensePolicyService) Create(ctx context.Context, req *model.ExpensePolicyRequest) (*model.ExpensePolicy, error) {
	policy := &model.ExpensePolicy{
		Category:             model.ExpenseCategory(req.Category),
		MonthlyLimit:         req.MonthlyLimit,
		PerClaimLimit:        req.PerClaimLimit,
		AutoApproveLimit:     req.AutoApproveLimit,
		RequiresReceiptAbove: req.RequiresReceiptAbove,
		Description:          req.Description,
		IsActive:             true,
	}
	if req.IsActive != nil {
		policy.IsActive = *req.IsActive
	}
	if err := s.deps.Repos.ExpensePolicy.Create(ctx, policy); err != nil {
		return nil, err
	}
	return policy, nil
}

func (s *expensePolicyService) Update(ctx context.Context, id uuid.UUID, req *model.ExpensePolicyRequest) (*model.ExpensePolicy, error) {
	policy, err := s.deps.Repos.ExpensePolicy.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	policy.Category = model.ExpenseCategory(req.Category)
	policy.MonthlyLimit = req.MonthlyLimit
	policy.PerClaimLimit = req.PerClaimLimit
	policy.AutoApproveLimit = req.AutoApproveLimit
	policy.RequiresReceiptAbove = req.RequiresReceiptAbove
	policy.Description = req.Description
	if req.IsActive != nil {
		policy.IsActive = *req.IsActive
	}
	if err := s.deps.Repos.ExpensePolicy.Update(ctx, policy); err != nil {
		return nil, err
	}
	return policy, nil
}

func (s *expensePolicyService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.deps.Repos.ExpensePolicy.Delete(ctx, id)
}

func (s *expensePolicyService) GetBudgets(ctx context.Context) ([]model.ExpenseBudget, error) {
	return s.deps.Repos.ExpenseBudget.FindAll(ctx)
}

func (s *expensePolicyService) GetPolicyViolations(ctx context.Context, userID uuid.UUID) ([]model.ExpensePolicyViolation, error) {
	return s.deps.Repos.ExpensePolicyViolation.FindByUserID(ctx, userID)
}

// ===== ExpenseNotificationService =====

type ExpenseNotificationService interface {
	GetNotifications(ctx context.Context, userID uuid.UUID, filter string) ([]model.ExpenseNotification, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	GetReminders(ctx context.Context, userID uuid.UUID) ([]model.ExpenseReminder, error)
	DismissReminder(ctx context.Context, id uuid.UUID) error
	GetSettings(ctx context.Context, userID uuid.UUID) (*model.ExpenseNotificationSetting, error)
	UpdateSettings(ctx context.Context, userID uuid.UUID, req *model.ExpenseNotificationSettingRequest) (*model.ExpenseNotificationSetting, error)
}

type expenseNotificationService struct {
	deps Deps
}

func NewExpenseNotificationService(deps Deps) ExpenseNotificationService {
	return &expenseNotificationService{deps: deps}
}

func (s *expenseNotificationService) GetNotifications(ctx context.Context, userID uuid.UUID, filter string) ([]model.ExpenseNotification, error) {
	return s.deps.Repos.ExpenseNotification.FindByUserID(ctx, userID, filter)
}

func (s *expenseNotificationService) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	return s.deps.Repos.ExpenseNotification.MarkAsRead(ctx, id)
}

func (s *expenseNotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return s.deps.Repos.ExpenseNotification.MarkAllAsRead(ctx, userID)
}

func (s *expenseNotificationService) GetReminders(ctx context.Context, userID uuid.UUID) ([]model.ExpenseReminder, error) {
	return s.deps.Repos.ExpenseReminder.FindByUserID(ctx, userID)
}

func (s *expenseNotificationService) DismissReminder(ctx context.Context, id uuid.UUID) error {
	return s.deps.Repos.ExpenseReminder.Dismiss(ctx, id)
}

func (s *expenseNotificationService) GetSettings(ctx context.Context, userID uuid.UUID) (*model.ExpenseNotificationSetting, error) {
	return s.deps.Repos.ExpenseNotificationSetting.FindByUserID(ctx, userID)
}

func (s *expenseNotificationService) UpdateSettings(ctx context.Context, userID uuid.UUID, req *model.ExpenseNotificationSettingRequest) (*model.ExpenseNotificationSetting, error) {
	setting, err := s.deps.Repos.ExpenseNotificationSetting.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	setting.UserID = userID
	if req.EmailEnabled != nil {
		setting.EmailEnabled = *req.EmailEnabled
	}
	if req.PushEnabled != nil {
		setting.PushEnabled = *req.PushEnabled
	}
	if req.ApprovalAlerts != nil {
		setting.ApprovalAlerts = *req.ApprovalAlerts
	}
	if req.ReimbursementAlerts != nil {
		setting.ReimbursementAlerts = *req.ReimbursementAlerts
	}
	if req.PolicyAlerts != nil {
		setting.PolicyAlerts = *req.PolicyAlerts
	}
	if req.WeeklyDigest != nil {
		setting.WeeklyDigest = *req.WeeklyDigest
	}
	if err := s.deps.Repos.ExpenseNotificationSetting.Upsert(ctx, setting); err != nil {
		return nil, err
	}
	return setting, nil
}

// ===== ExpenseApprovalFlowService =====

type ExpenseApprovalFlowService interface {
	GetConfig(ctx context.Context) (*model.ExpenseApprovalFlow, error)
	GetDelegates(ctx context.Context, userID uuid.UUID) ([]model.ExpenseDelegate, error)
	SetDelegate(ctx context.Context, userID uuid.UUID, req *model.ExpenseDelegateRequest) (*model.ExpenseDelegate, error)
	RemoveDelegate(ctx context.Context, id uuid.UUID) error
}

type expenseApprovalFlowService struct {
	deps Deps
}

func NewExpenseApprovalFlowService(deps Deps) ExpenseApprovalFlowService {
	return &expenseApprovalFlowService{deps: deps}
}

func (s *expenseApprovalFlowService) GetConfig(ctx context.Context) (*model.ExpenseApprovalFlow, error) {
	return s.deps.Repos.ExpenseApprovalFlow.FindActive(ctx)
}

func (s *expenseApprovalFlowService) GetDelegates(ctx context.Context, userID uuid.UUID) ([]model.ExpenseDelegate, error) {
	return s.deps.Repos.ExpenseDelegate.FindByUserID(ctx, userID)
}

func (s *expenseApprovalFlowService) SetDelegate(ctx context.Context, userID uuid.UUID, req *model.ExpenseDelegateRequest) (*model.ExpenseDelegate, error) {
	delegateID, err := uuid.Parse(req.DelegateTo)
	if err != nil {
		return nil, fmt.Errorf("無効な代理承認者IDです")
	}
	startDate, _ := time.Parse("2006-01-02", req.StartDate)
	endDate, _ := time.Parse("2006-01-02", req.EndDate)

	delegate := &model.ExpenseDelegate{
		UserID:     userID,
		DelegateID: delegateID,
		StartDate:  startDate,
		EndDate:    endDate,
		IsActive:   true,
	}
	if err := s.deps.Repos.ExpenseDelegate.Create(ctx, delegate); err != nil {
		return nil, err
	}
	return delegate, nil
}

func (s *expenseApprovalFlowService) RemoveDelegate(ctx context.Context, id uuid.UUID) error {
	return s.deps.Repos.ExpenseDelegate.Delete(ctx, id)
}
