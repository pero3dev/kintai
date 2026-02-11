package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/kintai/backend/internal/mocks"
	"github.com/your-org/kintai/backend/internal/model"
	"github.com/your-org/kintai/backend/internal/repository"
)

var errHRNotFound = errors.New("not found")

func hrStringPtr(v string) *string     { return &v }
func hrFloatPtr(v float64) *float64    { return &v }
func hrIntPtr(v int) *int              { return &v }
func hrBoolPtr(v bool) *bool           { return &v }
func hrUUIDPtr(v uuid.UUID) *uuid.UUID { return &v }

type hrServiceTestRepos struct {
	hrEmployee   *mockHREmployeeRepo
	hrDepartment *mockHRDepartmentRepo
	evaluation   *mockEvaluationRepo
	goal         *mockGoalRepo
	training     *mockTrainingRepo
	recruitment  *mockRecruitmentRepo
	document     *mockDocumentRepo
	announcement *mockAnnouncementRepo
	oneOnOne     *mockOneOnOneRepo
	skill        *mockSkillRepo
	salary       *mockSalaryRepo
	onboarding   *mockOnboardingRepo
	offboarding  *mockOffboardingRepo
	survey       *mockSurveyRepo

	user       *mocks.MockUserRepository
	attendance *mocks.MockAttendanceRepository
	leave      *mocks.MockLeaveRequestRepository
	overtime   *mockOvertimeRequestRepo
}

func setupHRServiceDeps(t *testing.T) (Deps, *hrServiceTestRepos) {
	t.Helper()
	deps, otRepo, userRepo := setupOvertimeDeps(t)
	r := &hrServiceTestRepos{
		hrEmployee:   newMockHREmployeeRepo(),
		hrDepartment: newMockHRDepartmentRepo(),
		evaluation:   newMockEvaluationRepo(),
		goal:         newMockGoalRepo(),
		training:     newMockTrainingRepo(),
		recruitment:  newMockRecruitmentRepo(),
		document:     newMockDocumentRepo(),
		announcement: newMockAnnouncementRepo(),
		oneOnOne:     newMockOneOnOneRepo(),
		skill:        newMockSkillRepo(),
		salary:       newMockSalaryRepo(),
		onboarding:   newMockOnboardingRepo(),
		offboarding:  newMockOffboardingRepo(),
		survey:       newMockSurveyRepo(),
		user:         userRepo,
		attendance:   deps.Repos.Attendance.(*mocks.MockAttendanceRepository),
		leave:        deps.Repos.LeaveRequest.(*mocks.MockLeaveRequestRepository),
		overtime:     otRepo,
	}

	deps.Repos.HREmployee = r.hrEmployee
	deps.Repos.HRDepartment = r.hrDepartment
	deps.Repos.Evaluation = r.evaluation
	deps.Repos.Goal = r.goal
	deps.Repos.Training = r.training
	deps.Repos.Recruitment = r.recruitment
	deps.Repos.Document = r.document
	deps.Repos.Announcement = r.announcement
	deps.Repos.OneOnOne = r.oneOnOne
	deps.Repos.Skill = r.skill
	deps.Repos.Salary = r.salary
	deps.Repos.Onboarding = r.onboarding
	deps.Repos.Offboarding = r.offboarding
	deps.Repos.Survey = r.survey

	return deps, r
}

type mockHREmployeeRepo struct {
	items         map[uuid.UUID]*model.HREmployee
	createErr     error
	findByIDErr   error
	findAllErr    error
	updateErr     error
	deleteErr     error
	findByDeptErr error
	active        int64
	total         int64
	useCount      bool
}

func newMockHREmployeeRepo() *mockHREmployeeRepo {
	return &mockHREmployeeRepo{items: map[uuid.UUID]*model.HREmployee{}}
}

func (m *mockHREmployeeRepo) Create(ctx context.Context, e *model.HREmployee) error {
	if m.createErr != nil {
		return m.createErr
	}
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	m.items[e.ID] = e
	return nil
}

func (m *mockHREmployeeRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.HREmployee, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	e, ok := m.items[id]
	if !ok {
		return nil, errHRNotFound
	}
	return e, nil
}

func (m *mockHREmployeeRepo) FindAll(ctx context.Context, page, pageSize int, department, status, employmentType, search string) ([]model.HREmployee, int64, error) {
	if m.findAllErr != nil {
		return nil, 0, m.findAllErr
	}
	var list []model.HREmployee
	for _, e := range m.items {
		if status != "" && string(e.Status) != status {
			continue
		}
		if employmentType != "" && string(e.EmploymentType) != employmentType {
			continue
		}
		if search != "" {
			key := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(e.FirstName), key) &&
				!strings.Contains(strings.ToLower(e.LastName), key) &&
				!strings.Contains(strings.ToLower(e.Email), key) {
				continue
			}
		}
		list = append(list, *e)
	}
	total := int64(len(list))
	if pageSize <= 0 || page <= 0 {
		return list, total, nil
	}
	start := (page - 1) * pageSize
	if start >= len(list) {
		return []model.HREmployee{}, total, nil
	}
	end := start + pageSize
	if end > len(list) {
		end = len(list)
	}
	return list[start:end], total, nil
}

func (m *mockHREmployeeRepo) Update(ctx context.Context, e *model.HREmployee) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.items[e.ID] = e
	return nil
}

func (m *mockHREmployeeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.items, id)
	return nil
}

func (m *mockHREmployeeRepo) FindByDepartmentID(ctx context.Context, deptID uuid.UUID) ([]model.HREmployee, error) {
	if m.findByDeptErr != nil {
		return nil, m.findByDeptErr
	}
	var out []model.HREmployee
	for _, e := range m.items {
		if e.DepartmentID != nil && *e.DepartmentID == deptID {
			out = append(out, *e)
		}
	}
	return out, nil
}

func (m *mockHREmployeeRepo) CountByStatus(ctx context.Context) (active int64, total int64, err error) {
	if m.useCount {
		return m.active, m.total, nil
	}
	var a int64
	for _, e := range m.items {
		if e.Status == model.EmployeeStatusActive {
			a++
		}
	}
	return a, int64(len(m.items)), nil
}

var _ repository.HREmployeeRepository = (*mockHREmployeeRepo)(nil)

type mockHRDepartmentRepo struct {
	items        map[uuid.UUID]*model.HRDepartment
	createErr    error
	findByIDErr  error
	findAllErr   error
	updateErr    error
	deleteErr    error
}

func newMockHRDepartmentRepo() *mockHRDepartmentRepo {
	return &mockHRDepartmentRepo{items: map[uuid.UUID]*model.HRDepartment{}}
}

func (m *mockHRDepartmentRepo) Create(ctx context.Context, d *model.HRDepartment) error {
	if m.createErr != nil {
		return m.createErr
	}
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	m.items[d.ID] = d
	return nil
}
func (m *mockHRDepartmentRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.HRDepartment, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	d, ok := m.items[id]
	if !ok {
		return nil, errHRNotFound
	}
	return d, nil
}
func (m *mockHRDepartmentRepo) FindAll(ctx context.Context) ([]model.HRDepartment, error) {
	if m.findAllErr != nil {
		return nil, m.findAllErr
	}
	var out []model.HRDepartment
	for _, d := range m.items {
		out = append(out, *d)
	}
	return out, nil
}
func (m *mockHRDepartmentRepo) Update(ctx context.Context, d *model.HRDepartment) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.items[d.ID] = d
	return nil
}
func (m *mockHRDepartmentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.items, id)
	return nil
}

var _ repository.HRDepartmentRepository = (*mockHRDepartmentRepo)(nil)

type mockEvaluationRepo struct {
	evals          map[uuid.UUID]*model.Evaluation
	cycles         map[uuid.UUID]*model.EvaluationCycle
	createErr      error
	findByIDErr    error
	findAllErr     error
	updateErr      error
	createCycleErr error
	findCyclesErr  error
}

func newMockEvaluationRepo() *mockEvaluationRepo {
	return &mockEvaluationRepo{
		evals:  map[uuid.UUID]*model.Evaluation{},
		cycles: map[uuid.UUID]*model.EvaluationCycle{},
	}
}
func (m *mockEvaluationRepo) Create(ctx context.Context, e *model.Evaluation) error {
	if m.createErr != nil {
		return m.createErr
	}
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	m.evals[e.ID] = e
	return nil
}
func (m *mockEvaluationRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Evaluation, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	e, ok := m.evals[id]
	if !ok {
		return nil, errHRNotFound
	}
	return e, nil
}
func (m *mockEvaluationRepo) FindAll(ctx context.Context, page, pageSize int, cycleID, status string) ([]model.Evaluation, int64, error) {
	if m.findAllErr != nil {
		return nil, 0, m.findAllErr
	}
	var out []model.Evaluation
	for _, e := range m.evals {
		if cycleID != "" && e.CycleID.String() != cycleID {
			continue
		}
		if status != "" && string(e.Status) != status {
			continue
		}
		out = append(out, *e)
	}
	return out, int64(len(out)), nil
}
func (m *mockEvaluationRepo) Update(ctx context.Context, e *model.Evaluation) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.evals[e.ID] = e
	return nil
}
func (m *mockEvaluationRepo) CreateCycle(ctx context.Context, c *model.EvaluationCycle) error {
	if m.createCycleErr != nil {
		return m.createCycleErr
	}
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	m.cycles[c.ID] = c
	return nil
}
func (m *mockEvaluationRepo) FindAllCycles(ctx context.Context) ([]model.EvaluationCycle, error) {
	if m.findCyclesErr != nil {
		return nil, m.findCyclesErr
	}
	var out []model.EvaluationCycle
	for _, c := range m.cycles {
		out = append(out, *c)
	}
	return out, nil
}

var _ repository.EvaluationRepository = (*mockEvaluationRepo)(nil)

type mockGoalRepo struct {
	items       map[uuid.UUID]*model.HRGoal
	createErr   error
	findByIDErr error
	findAllErr  error
	updateErr   error
	deleteErr   error
}

func newMockGoalRepo() *mockGoalRepo { return &mockGoalRepo{items: map[uuid.UUID]*model.HRGoal{}} }
func (m *mockGoalRepo) Create(ctx context.Context, g *model.HRGoal) error {
	if m.createErr != nil {
		return m.createErr
	}
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	m.items[g.ID] = g
	return nil
}
func (m *mockGoalRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.HRGoal, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	g, ok := m.items[id]
	if !ok {
		return nil, errHRNotFound
	}
	return g, nil
}
func (m *mockGoalRepo) FindAll(ctx context.Context, page, pageSize int, status, category, employeeID string) ([]model.HRGoal, int64, error) {
	if m.findAllErr != nil {
		return nil, 0, m.findAllErr
	}
	var out []model.HRGoal
	for _, g := range m.items {
		if status != "" && string(g.Status) != status {
			continue
		}
		if category != "" && string(g.Category) != category {
			continue
		}
		if employeeID != "" && g.EmployeeID.String() != employeeID {
			continue
		}
		out = append(out, *g)
	}
	return out, int64(len(out)), nil
}
func (m *mockGoalRepo) Update(ctx context.Context, g *model.HRGoal) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.items[g.ID] = g
	return nil
}
func (m *mockGoalRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.items, id)
	return nil
}

var _ repository.GoalRepository = (*mockGoalRepo)(nil)

type mockTrainingRepo struct {
	programs          map[uuid.UUID]*model.TrainingProgram
	enrollments       map[string]*model.TrainingEnrollment
	createErr         error
	findByIDErr       error
	findAllErr        error
	updateErr         error
	deleteErr         error
	createEnrollErr   error
	updateEnrollErr   error
	findEnrollmentErr error
}

func newMockTrainingRepo() *mockTrainingRepo {
	return &mockTrainingRepo{
		programs:    map[uuid.UUID]*model.TrainingProgram{},
		enrollments: map[string]*model.TrainingEnrollment{},
	}
}

func enrollmentKey(programID, employeeID uuid.UUID) string {
	return programID.String() + ":" + employeeID.String()
}

func (m *mockTrainingRepo) Create(ctx context.Context, t *model.TrainingProgram) error {
	if m.createErr != nil {
		return m.createErr
	}
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	m.programs[t.ID] = t
	return nil
}
func (m *mockTrainingRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.TrainingProgram, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	v, ok := m.programs[id]
	if !ok {
		return nil, errHRNotFound
	}
	return v, nil
}
func (m *mockTrainingRepo) FindAll(ctx context.Context, page, pageSize int, category, status string) ([]model.TrainingProgram, int64, error) {
	if m.findAllErr != nil {
		return nil, 0, m.findAllErr
	}
	var out []model.TrainingProgram
	for _, p := range m.programs {
		if category != "" && p.Category != category {
			continue
		}
		if status != "" && string(p.Status) != status {
			continue
		}
		out = append(out, *p)
	}
	return out, int64(len(out)), nil
}
func (m *mockTrainingRepo) Update(ctx context.Context, t *model.TrainingProgram) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.programs[t.ID] = t
	return nil
}
func (m *mockTrainingRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.programs, id)
	return nil
}
func (m *mockTrainingRepo) CreateEnrollment(ctx context.Context, e *model.TrainingEnrollment) error {
	if m.createEnrollErr != nil {
		return m.createEnrollErr
	}
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	m.enrollments[enrollmentKey(e.ProgramID, e.EmployeeID)] = e
	return nil
}
func (m *mockTrainingRepo) UpdateEnrollment(ctx context.Context, e *model.TrainingEnrollment) error {
	if m.updateEnrollErr != nil {
		return m.updateEnrollErr
	}
	m.enrollments[enrollmentKey(e.ProgramID, e.EmployeeID)] = e
	return nil
}
func (m *mockTrainingRepo) FindEnrollment(ctx context.Context, programID, employeeID uuid.UUID) (*model.TrainingEnrollment, error) {
	if m.findEnrollmentErr != nil {
		return nil, m.findEnrollmentErr
	}
	v, ok := m.enrollments[enrollmentKey(programID, employeeID)]
	if !ok {
		return nil, errHRNotFound
	}
	return v, nil
}

var _ repository.TrainingRepository = (*mockTrainingRepo)(nil)

type mockRecruitmentRepo struct {
	positions          map[uuid.UUID]*model.RecruitmentPosition
	applicants         map[uuid.UUID]*model.Applicant
	createPositionErr  error
	findPositionErr    error
	findAllPositionErr error
	updatePositionErr  error
	createApplicantErr error
	findApplicantsErr  error
	findApplicantErr   error
	updateApplicantErr error
}

func newMockRecruitmentRepo() *mockRecruitmentRepo {
	return &mockRecruitmentRepo{
		positions:  map[uuid.UUID]*model.RecruitmentPosition{},
		applicants: map[uuid.UUID]*model.Applicant{},
	}
}
func (m *mockRecruitmentRepo) CreatePosition(ctx context.Context, p *model.RecruitmentPosition) error {
	if m.createPositionErr != nil {
		return m.createPositionErr
	}
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	m.positions[p.ID] = p
	return nil
}
func (m *mockRecruitmentRepo) FindPositionByID(ctx context.Context, id uuid.UUID) (*model.RecruitmentPosition, error) {
	if m.findPositionErr != nil {
		return nil, m.findPositionErr
	}
	v, ok := m.positions[id]
	if !ok {
		return nil, errHRNotFound
	}
	return v, nil
}
func (m *mockRecruitmentRepo) FindAllPositions(ctx context.Context, page, pageSize int, status, department string) ([]model.RecruitmentPosition, int64, error) {
	if m.findAllPositionErr != nil {
		return nil, 0, m.findAllPositionErr
	}
	var out []model.RecruitmentPosition
	for _, p := range m.positions {
		if status != "" && string(p.Status) != status {
			continue
		}
		out = append(out, *p)
	}
	return out, int64(len(out)), nil
}
func (m *mockRecruitmentRepo) UpdatePosition(ctx context.Context, p *model.RecruitmentPosition) error {
	if m.updatePositionErr != nil {
		return m.updatePositionErr
	}
	m.positions[p.ID] = p
	return nil
}
func (m *mockRecruitmentRepo) CreateApplicant(ctx context.Context, a *model.Applicant) error {
	if m.createApplicantErr != nil {
		return m.createApplicantErr
	}
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	m.applicants[a.ID] = a
	return nil
}
func (m *mockRecruitmentRepo) FindAllApplicants(ctx context.Context, positionID, stage string) ([]model.Applicant, error) {
	if m.findApplicantsErr != nil {
		return nil, m.findApplicantsErr
	}
	var out []model.Applicant
	for _, a := range m.applicants {
		if positionID != "" && a.PositionID.String() != positionID {
			continue
		}
		if stage != "" && string(a.Stage) != stage {
			continue
		}
		out = append(out, *a)
	}
	return out, nil
}
func (m *mockRecruitmentRepo) FindApplicantByID(ctx context.Context, id uuid.UUID) (*model.Applicant, error) {
	if m.findApplicantErr != nil {
		return nil, m.findApplicantErr
	}
	v, ok := m.applicants[id]
	if !ok {
		return nil, errHRNotFound
	}
	return v, nil
}
func (m *mockRecruitmentRepo) UpdateApplicant(ctx context.Context, a *model.Applicant) error {
	if m.updateApplicantErr != nil {
		return m.updateApplicantErr
	}
	m.applicants[a.ID] = a
	return nil
}

var _ repository.RecruitmentRepository = (*mockRecruitmentRepo)(nil)

type mockDocumentRepo struct {
	items       map[uuid.UUID]*model.HRDocument
	createErr   error
	findAllErr  error
	findByIDErr error
	deleteErr   error
}

func newMockDocumentRepo() *mockDocumentRepo { return &mockDocumentRepo{items: map[uuid.UUID]*model.HRDocument{}} }
func (m *mockDocumentRepo) Create(ctx context.Context, d *model.HRDocument) error {
	if m.createErr != nil {
		return m.createErr
	}
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	m.items[d.ID] = d
	return nil
}
func (m *mockDocumentRepo) FindAll(ctx context.Context, page, pageSize int, docType, employeeID string) ([]model.HRDocument, int64, error) {
	if m.findAllErr != nil {
		return nil, 0, m.findAllErr
	}
	var out []model.HRDocument
	for _, d := range m.items {
		if docType != "" && d.Type != docType {
			continue
		}
		if employeeID != "" {
			if d.EmployeeID == nil || d.EmployeeID.String() != employeeID {
				continue
			}
		}
		out = append(out, *d)
	}
	return out, int64(len(out)), nil
}
func (m *mockDocumentRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.HRDocument, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	v, ok := m.items[id]
	if !ok {
		return nil, errHRNotFound
	}
	return v, nil
}
func (m *mockDocumentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.items, id)
	return nil
}

var _ repository.DocumentRepository = (*mockDocumentRepo)(nil)

type mockAnnouncementRepo struct {
	items       map[uuid.UUID]*model.HRAnnouncement
	createErr   error
	findByIDErr error
	findAllErr  error
	updateErr   error
	deleteErr   error
}

func newMockAnnouncementRepo() *mockAnnouncementRepo { return &mockAnnouncementRepo{items: map[uuid.UUID]*model.HRAnnouncement{}} }
func (m *mockAnnouncementRepo) Create(ctx context.Context, a *model.HRAnnouncement) error {
	if m.createErr != nil {
		return m.createErr
	}
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	m.items[a.ID] = a
	return nil
}
func (m *mockAnnouncementRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.HRAnnouncement, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	v, ok := m.items[id]
	if !ok {
		return nil, errHRNotFound
	}
	return v, nil
}
func (m *mockAnnouncementRepo) FindAll(ctx context.Context, page, pageSize int, priority string) ([]model.HRAnnouncement, int64, error) {
	if m.findAllErr != nil {
		return nil, 0, m.findAllErr
	}
	var out []model.HRAnnouncement
	for _, a := range m.items {
		if priority != "" && string(a.Priority) != priority {
			continue
		}
		out = append(out, *a)
	}
	return out, int64(len(out)), nil
}
func (m *mockAnnouncementRepo) Update(ctx context.Context, a *model.HRAnnouncement) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.items[a.ID] = a
	return nil
}
func (m *mockAnnouncementRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.items, id)
	return nil
}

var _ repository.AnnouncementRepository = (*mockAnnouncementRepo)(nil)

type mockOneOnOneRepo struct {
	items       map[uuid.UUID]*model.OneOnOneMeeting
	createErr   error
	findByIDErr error
	findAllErr  error
	updateErr   error
	deleteErr   error
}

func newMockOneOnOneRepo() *mockOneOnOneRepo { return &mockOneOnOneRepo{items: map[uuid.UUID]*model.OneOnOneMeeting{}} }
func (m *mockOneOnOneRepo) Create(ctx context.Context, meeting *model.OneOnOneMeeting) error {
	if m.createErr != nil {
		return m.createErr
	}
	if meeting.ID == uuid.Nil {
		meeting.ID = uuid.New()
	}
	m.items[meeting.ID] = meeting
	return nil
}
func (m *mockOneOnOneRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.OneOnOneMeeting, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	v, ok := m.items[id]
	if !ok {
		return nil, errHRNotFound
	}
	return v, nil
}
func (m *mockOneOnOneRepo) FindAll(ctx context.Context, status, employeeID string) ([]model.OneOnOneMeeting, error) {
	if m.findAllErr != nil {
		return nil, m.findAllErr
	}
	var out []model.OneOnOneMeeting
	for _, v := range m.items {
		if status != "" && v.Status != status {
			continue
		}
		if employeeID != "" && v.EmployeeID.String() != employeeID && v.ManagerID.String() != employeeID {
			continue
		}
		out = append(out, *v)
	}
	return out, nil
}
func (m *mockOneOnOneRepo) Update(ctx context.Context, meeting *model.OneOnOneMeeting) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.items[meeting.ID] = meeting
	return nil
}
func (m *mockOneOnOneRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.items, id)
	return nil
}

var _ repository.OneOnOneRepository = (*mockOneOnOneRepo)(nil)

type mockSkillRepo struct {
	items           map[uuid.UUID]*model.EmployeeSkill
	createErr       error
	findByIDErr     error
	updateErr       error
	findAllErr      error
	findEmployeeErr error
	gapAnalysis     []map[string]interface{}
	gapErr          error
}

func newMockSkillRepo() *mockSkillRepo { return &mockSkillRepo{items: map[uuid.UUID]*model.EmployeeSkill{}} }
func (m *mockSkillRepo) Create(ctx context.Context, s *model.EmployeeSkill) error {
	if m.createErr != nil {
		return m.createErr
	}
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	m.items[s.ID] = s
	return nil
}
func (m *mockSkillRepo) FindByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]model.EmployeeSkill, error) {
	if m.findEmployeeErr != nil {
		return nil, m.findEmployeeErr
	}
	var out []model.EmployeeSkill
	for _, s := range m.items {
		if s.EmployeeID == employeeID {
			out = append(out, *s)
		}
	}
	return out, nil
}
func (m *mockSkillRepo) FindAll(ctx context.Context, department string) ([]model.EmployeeSkill, error) {
	if m.findAllErr != nil {
		return nil, m.findAllErr
	}
	var out []model.EmployeeSkill
	for _, s := range m.items {
		out = append(out, *s)
	}
	return out, nil
}
func (m *mockSkillRepo) Update(ctx context.Context, s *model.EmployeeSkill) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.items[s.ID] = s
	return nil
}
func (m *mockSkillRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.EmployeeSkill, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	v, ok := m.items[id]
	if !ok {
		return nil, errHRNotFound
	}
	return v, nil
}
func (m *mockSkillRepo) GetGapAnalysis(ctx context.Context, department string) ([]map[string]interface{}, error) {
	if m.gapErr != nil {
		return nil, m.gapErr
	}
	return m.gapAnalysis, nil
}

var _ repository.SkillRepository = (*mockSkillRepo)(nil)

type mockSalaryRepo struct {
	records     map[uuid.UUID][]model.SalaryRecord
	overview    map[string]interface{}
	findByIDErr error
	createErr   error
	overviewErr error
}

func newMockSalaryRepo() *mockSalaryRepo {
	return &mockSalaryRepo{
		records:  map[uuid.UUID][]model.SalaryRecord{},
		overview: map[string]interface{}{"total_payroll": 0.0, "headcount": int64(0)},
	}
}
func (m *mockSalaryRepo) FindByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]model.SalaryRecord, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	return m.records[employeeID], nil
}
func (m *mockSalaryRepo) Create(ctx context.Context, s *model.SalaryRecord) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.records[s.EmployeeID] = append(m.records[s.EmployeeID], *s)
	return nil
}
func (m *mockSalaryRepo) GetOverview(ctx context.Context, department string) (map[string]interface{}, error) {
	if m.overviewErr != nil {
		return nil, m.overviewErr
	}
	return m.overview, nil
}

var _ repository.SalaryRepository = (*mockSalaryRepo)(nil)

type mockOnboardingRepo struct {
	items             map[uuid.UUID]*model.Onboarding
	templates         map[uuid.UUID]*model.OnboardingTemplate
	createErr         error
	findByIDErr       error
	findAllErr        error
	updateErr         error
	createTemplateErr error
	findTemplatesErr  error
}

func newMockOnboardingRepo() *mockOnboardingRepo {
	return &mockOnboardingRepo{
		items:     map[uuid.UUID]*model.Onboarding{},
		templates: map[uuid.UUID]*model.OnboardingTemplate{},
	}
}
func (m *mockOnboardingRepo) Create(ctx context.Context, o *model.Onboarding) error {
	if m.createErr != nil {
		return m.createErr
	}
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	m.items[o.ID] = o
	return nil
}
func (m *mockOnboardingRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Onboarding, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	v, ok := m.items[id]
	if !ok {
		return nil, errHRNotFound
	}
	return v, nil
}
func (m *mockOnboardingRepo) FindAll(ctx context.Context, status string) ([]model.Onboarding, error) {
	if m.findAllErr != nil {
		return nil, m.findAllErr
	}
	var out []model.Onboarding
	for _, o := range m.items {
		if status != "" && string(o.Status) != status {
			continue
		}
		out = append(out, *o)
	}
	return out, nil
}
func (m *mockOnboardingRepo) Update(ctx context.Context, o *model.Onboarding) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.items[o.ID] = o
	return nil
}
func (m *mockOnboardingRepo) CreateTemplate(ctx context.Context, t *model.OnboardingTemplate) error {
	if m.createTemplateErr != nil {
		return m.createTemplateErr
	}
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	m.templates[t.ID] = t
	return nil
}
func (m *mockOnboardingRepo) FindAllTemplates(ctx context.Context) ([]model.OnboardingTemplate, error) {
	if m.findTemplatesErr != nil {
		return nil, m.findTemplatesErr
	}
	var out []model.OnboardingTemplate
	for _, t := range m.templates {
		out = append(out, *t)
	}
	return out, nil
}

var _ repository.OnboardingRepository = (*mockOnboardingRepo)(nil)

type mockOffboardingRepo struct {
	items        map[uuid.UUID]*model.Offboarding
	createErr    error
	findByIDErr  error
	findAllErr   error
	updateErr    error
	analyticsErr error
	analytics    map[string]interface{}
}

func newMockOffboardingRepo() *mockOffboardingRepo {
	return &mockOffboardingRepo{
		items:     map[uuid.UUID]*model.Offboarding{},
		analytics: map[string]interface{}{"turnover_rate": 1.0},
	}
}
func (m *mockOffboardingRepo) Create(ctx context.Context, o *model.Offboarding) error {
	if m.createErr != nil {
		return m.createErr
	}
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	m.items[o.ID] = o
	return nil
}
func (m *mockOffboardingRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Offboarding, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	v, ok := m.items[id]
	if !ok {
		return nil, errHRNotFound
	}
	return v, nil
}
func (m *mockOffboardingRepo) FindAll(ctx context.Context, status string) ([]model.Offboarding, error) {
	if m.findAllErr != nil {
		return nil, m.findAllErr
	}
	var out []model.Offboarding
	for _, o := range m.items {
		if status != "" && string(o.Status) != status {
			continue
		}
		out = append(out, *o)
	}
	return out, nil
}
func (m *mockOffboardingRepo) Update(ctx context.Context, o *model.Offboarding) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.items[o.ID] = o
	return nil
}
func (m *mockOffboardingRepo) GetTurnoverAnalytics(ctx context.Context) (map[string]interface{}, error) {
	if m.analyticsErr != nil {
		return nil, m.analyticsErr
	}
	return m.analytics, nil
}

var _ repository.OffboardingRepository = (*mockOffboardingRepo)(nil)

type mockSurveyRepo struct {
	surveys           map[uuid.UUID]*model.Survey
	responses         map[uuid.UUID][]model.SurveyResponse
	createErr         error
	findByIDErr       error
	findAllErr        error
	updateErr         error
	deleteErr         error
	createResponseErr error
	findResponsesErr  error
	countResponsesErr error
}

func newMockSurveyRepo() *mockSurveyRepo {
	return &mockSurveyRepo{
		surveys:   map[uuid.UUID]*model.Survey{},
		responses: map[uuid.UUID][]model.SurveyResponse{},
	}
}
func (m *mockSurveyRepo) Create(ctx context.Context, s *model.Survey) error {
	if m.createErr != nil {
		return m.createErr
	}
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	m.surveys[s.ID] = s
	return nil
}
func (m *mockSurveyRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Survey, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	v, ok := m.surveys[id]
	if !ok {
		return nil, errHRNotFound
	}
	return v, nil
}
func (m *mockSurveyRepo) FindAll(ctx context.Context, status, surveyType string) ([]model.Survey, error) {
	if m.findAllErr != nil {
		return nil, m.findAllErr
	}
	var out []model.Survey
	for _, s := range m.surveys {
		if status != "" && string(s.Status) != status {
			continue
		}
		if surveyType != "" && s.Type != surveyType {
			continue
		}
		out = append(out, *s)
	}
	return out, nil
}
func (m *mockSurveyRepo) Update(ctx context.Context, s *model.Survey) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.surveys[s.ID] = s
	return nil
}
func (m *mockSurveyRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.surveys, id)
	return nil
}
func (m *mockSurveyRepo) CreateResponse(ctx context.Context, r *model.SurveyResponse) error {
	if m.createResponseErr != nil {
		return m.createResponseErr
	}
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	m.responses[r.SurveyID] = append(m.responses[r.SurveyID], *r)
	return nil
}
func (m *mockSurveyRepo) FindResponsesBySurveyID(ctx context.Context, surveyID uuid.UUID) ([]model.SurveyResponse, error) {
	if m.findResponsesErr != nil {
		return nil, m.findResponsesErr
	}
	return m.responses[surveyID], nil
}
func (m *mockSurveyRepo) CountResponsesBySurveyID(ctx context.Context, surveyID uuid.UUID) (int64, error) {
	if m.countResponsesErr != nil {
		return 0, m.countResponsesErr
	}
	return int64(len(m.responses[surveyID])), nil
}

var _ repository.SurveyRepository = (*mockSurveyRepo)(nil)

func jsonUnmarshal(b []byte, v interface{}) error {
	if len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, v)
}

func TestHRService_HREmployeeAndDepartment(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()

	empSvc := NewHREmployeeService(deps)
	deptSvc := NewHRDepartmentService(deps)

	t.Run("employee_create_default_and_dates", func(t *testing.T) {
		got, err := empSvc.Create(ctx, model.HREmployeeCreateRequest{
			EmployeeCode: "E-001",
			FirstName:    "Taro",
			LastName:     "Test",
			Email:        "taro@test.local",
			HireDate:     "2024-01-10",
			BirthDate:    "1990-06-01",
		})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		if got.EmploymentType != model.EmploymentTypeFullTime {
			t.Fatalf("expected default employment type full_time, got %s", got.EmploymentType)
		}
		if got.HireDate == nil || got.BirthDate == nil {
			t.Fatalf("expected parsed dates")
		}
	})

	t.Run("employee_update_all_fields_and_errors", func(t *testing.T) {
		id := uuid.New()
		deptID := uuid.New()
		managerID := uuid.New()
		r.hrEmployee.items[id] = &model.HREmployee{BaseModel: model.BaseModel{ID: id}}

		firstName := "Hanako"
		lastName := "Updated"
		email := "hanako@test.local"
		phone := "090"
		position := "Engineer"
		grade := "M1"
		employmentType := "part_time"
		status := "inactive"
		address := "Tokyo"
		baseSalary := 450000.0

		updated, err := empSvc.Update(ctx, id, model.HREmployeeUpdateRequest{
			FirstName:      &firstName,
			LastName:       &lastName,
			Email:          &email,
			Phone:          &phone,
			Position:       &position,
			Grade:          &grade,
			DepartmentID:   &deptID,
			ManagerID:      &managerID,
			EmploymentType: &employmentType,
			Status:         &status,
			Address:        &address,
			BaseSalary:     &baseSalary,
		})
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}
		if updated.FirstName != firstName || updated.BaseSalary != baseSalary || updated.Status != model.EmployeeStatus(status) {
			t.Fatalf("update did not apply requested fields")
		}

		r.hrEmployee.findByIDErr = errors.New("find error")
		if _, err := empSvc.Update(ctx, uuid.New(), model.HREmployeeUpdateRequest{}); err == nil {
			t.Fatalf("expected find error")
		}
		r.hrEmployee.findByIDErr = nil
		r.hrEmployee.updateErr = errors.New("update error")
		if _, err := empSvc.Update(ctx, id, model.HREmployeeUpdateRequest{}); err == nil {
			t.Fatalf("expected update error")
		}
		r.hrEmployee.updateErr = nil
	})

	t.Run("employee_passthrough_and_create_error", func(t *testing.T) {
		r.hrEmployee.createErr = errors.New("create error")
		if _, err := empSvc.Create(ctx, model.HREmployeeCreateRequest{}); err == nil {
			t.Fatalf("expected create error")
		}
		r.hrEmployee.createErr = nil

		list, total, err := empSvc.FindAll(ctx, 1, 10, "", "", "", "")
		if err != nil || total == 0 || len(list) == 0 {
			t.Fatalf("FindAll failed: total=%d len=%d err=%v", total, len(list), err)
		}
		if err := empSvc.Delete(ctx, list[0].ID); err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
	})

	t.Run("department_update_and_errors", func(t *testing.T) {
		d, err := deptSvc.Create(ctx, model.HRDepartmentCreateRequest{Name: "Engineering", Code: "ENG"})
		if err != nil {
			t.Fatalf("Create department failed: %v", err)
		}
		newName := "Platform"
		newCode := "PLT"
		newDesc := "desc"
		newBudget := 1000000.0
		updated, err := deptSvc.Update(ctx, d.ID, model.HRDepartmentUpdateRequest{
			Name:        &newName,
			Code:        &newCode,
			Description: &newDesc,
			Budget:      &newBudget,
		})
		if err != nil {
			t.Fatalf("Update department failed: %v", err)
		}
		if updated.Name != "Platform" || updated.Budget != newBudget {
			t.Fatalf("department update not applied")
		}

		r.hrDepartment.findByIDErr = errors.New("find error")
		if _, err := deptSvc.Update(ctx, uuid.New(), model.HRDepartmentUpdateRequest{}); err == nil {
			t.Fatalf("expected find error")
		}
		r.hrDepartment.findByIDErr = nil
		r.hrDepartment.updateErr = errors.New("update error")
		if _, err := deptSvc.Update(ctx, d.ID, model.HRDepartmentUpdateRequest{}); err == nil {
			t.Fatalf("expected update error")
		}
		r.hrDepartment.updateErr = nil
	})
}

func TestHRService_EvaluationGoalTrainingRecruitment(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()

	evaluationSvc := NewEvaluationService(deps)
	goalSvc := NewGoalService(deps)
	trainingSvc := NewTrainingService(deps)
	recruitSvc := NewRecruitmentService(deps)

	t.Run("evaluation_update_submit_cycle", func(t *testing.T) {
		eid := uuid.New()
		ev := &model.Evaluation{BaseModel: model.BaseModel{ID: eid}, Status: model.EvaluationStatusDraft}
		r.evaluation.evals[eid] = ev

		score := 4.0
		comment := "ok"
		if _, err := evaluationSvc.Update(ctx, eid, model.EvaluationUpdateRequest{
			SelfScore:    &score,
			SelfComment:  &comment,
			ManagerScore: &score,
			FinalScore:   &score,
		}); err != nil {
			t.Fatalf("Update evaluation failed: %v", err)
		}
		submitted, err := evaluationSvc.Submit(ctx, eid)
		if err != nil || submitted.Status != model.EvaluationStatusSubmitted {
			t.Fatalf("Submit failed: %v status=%s", err, submitted.Status)
		}

		r.evaluation.updateErr = errors.New("update error")
		if _, err := evaluationSvc.Submit(ctx, eid); err == nil {
			t.Fatalf("expected submit update error")
		}
		r.evaluation.updateErr = nil
		r.evaluation.createCycleErr = errors.New("cycle error")
		if _, err := evaluationSvc.CreateCycle(ctx, model.EvaluationCycleCreateRequest{Name: "2024"}); err == nil {
			t.Fatalf("expected cycle create error")
		}
	})

	t.Run("goal_create_defaults_and_progress_branches", func(t *testing.T) {
		userID := uuid.New()
		g, err := goalSvc.Create(ctx, model.HRGoalCreateRequest{Title: "Goal"}, userID)
		if err != nil {
			t.Fatalf("Create goal failed: %v", err)
		}
		if g.Category != model.GoalCategoryPerformance || g.Weight != 1 {
			t.Fatalf("default category/weight not applied")
		}

		gid := g.ID
		r.goal.items[gid] = g
		if _, err := goalSvc.UpdateProgress(ctx, gid, 20); err != nil {
			t.Fatalf("progress update failed: %v", err)
		}
		if r.goal.items[gid].Status != model.GoalStatusInProgress {
			t.Fatalf("expected in_progress for progress=20")
		}
		if _, err := goalSvc.UpdateProgress(ctx, gid, 100); err != nil {
			t.Fatalf("progress update failed: %v", err)
		}
		if r.goal.items[gid].Status != model.GoalStatusCompleted {
			t.Fatalf("expected completed for progress=100")
		}
		r.goal.updateErr = errors.New("update error")
		if _, err := goalSvc.UpdateProgress(ctx, gid, 0); err == nil {
			t.Fatalf("expected progress update error")
		}
	})

	t.Run("training_complete_and_update_errors", func(t *testing.T) {
		pid := uuid.New()
		empID := uuid.New()
		r.training.programs[pid] = &model.TrainingProgram{BaseModel: model.BaseModel{ID: pid}, Title: "Go"}
		r.training.enrollments[enrollmentKey(pid, empID)] = &model.TrainingEnrollment{
			BaseModel:  model.BaseModel{ID: uuid.New()},
			ProgramID:  pid,
			EmployeeID: empID,
			Status:     "enrolled",
		}
		if err := trainingSvc.Complete(ctx, pid, empID); err != nil {
			t.Fatalf("Complete failed: %v", err)
		}
		if r.training.enrollments[enrollmentKey(pid, empID)].CompletedAt == nil {
			t.Fatalf("expected CompletedAt to be set")
		}

		r.training.findEnrollmentErr = errors.New("find enroll error")
		if err := trainingSvc.Complete(ctx, pid, empID); err == nil {
			t.Fatalf("expected complete find error")
		}
		r.training.findEnrollmentErr = nil
		r.training.createEnrollErr = errors.New("enroll error")
		if err := trainingSvc.Enroll(ctx, pid, empID); err == nil {
			t.Fatalf("expected enroll error")
		}
	})

	t.Run("recruitment_position_and_applicant", func(t *testing.T) {
		p, err := recruitSvc.CreatePosition(ctx, model.PositionCreateRequest{Title: "Engineer"})
		if err != nil {
			t.Fatalf("CreatePosition failed: %v", err)
		}
		if p.Openings != 1 {
			t.Fatalf("expected default openings=1, got %d", p.Openings)
		}

		app, err := recruitSvc.CreateApplicant(ctx, model.ApplicantCreateRequest{
			PositionID: p.ID,
			Name:       "Alice",
			Email:      "alice@example.com",
		})
		if err != nil {
			t.Fatalf("CreateApplicant failed: %v", err)
		}
		if app.Stage != model.ApplicantStageNew {
			t.Fatalf("expected default stage=new")
		}

		updated, err := recruitSvc.UpdateApplicantStage(ctx, app.ID, string(model.ApplicantStageInterview))
		if err != nil {
			t.Fatalf("UpdateApplicantStage failed: %v", err)
		}
		if updated.Stage != model.ApplicantStageInterview {
			t.Fatalf("stage not updated")
		}

		r.recruitment.findApplicantErr = errors.New("find applicant error")
		if _, err := recruitSvc.UpdateApplicantStage(ctx, app.ID, string(model.ApplicantStageOffer)); err == nil {
			t.Fatalf("expected update applicant find error")
		}
	})
}

func TestHRService_AnnouncementDashboardOrgChart(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()

	announcementSvc := NewAnnouncementService(deps)
	dashboardSvc := NewHRDashboardService(deps)
	orgSvc := NewOrgChartService(deps)

	t.Run("announcement_create_update", func(t *testing.T) {
		authorID := uuid.New()
		a, err := announcementSvc.Create(ctx, model.AnnouncementCreateRequest{
			Title: "Notice", Content: "content",
		}, authorID)
		if err != nil {
			t.Fatalf("Create announcement failed: %v", err)
		}
		if a.Priority != model.AnnouncementPriorityNormal {
			t.Fatalf("expected default priority normal")
		}

		title := "Updated"
		publish := true
		updated, err := announcementSvc.Update(ctx, a.ID, model.AnnouncementUpdateRequest{
			Title:       &title,
			IsPublished: &publish,
		})
		if err != nil {
			t.Fatalf("Update announcement failed: %v", err)
		}
		if !updated.IsPublished || updated.PublishedAt == nil {
			t.Fatalf("expected published=true and PublishedAt set")
		}

		publish = false
		updated, err = announcementSvc.Update(ctx, a.ID, model.AnnouncementUpdateRequest{
			IsPublished: &publish,
		})
		if err != nil {
			t.Fatalf("Update announcement failed: %v", err)
		}
		if updated.IsPublished {
			t.Fatalf("expected IsPublished=false")
		}
	})

	t.Run("dashboard_stats_and_recent_activities", func(t *testing.T) {
		now := time.Now()
		recentHire := now.AddDate(0, 0, -5)
		oldHire := now.AddDate(0, -6, 0)
		deptID := uuid.New()
		r.hrDepartment.items[deptID] = &model.HRDepartment{BaseModel: model.BaseModel{ID: deptID}, Name: "Dev"}
		r.hrDepartment.items[uuid.New()] = &model.HRDepartment{BaseModel: model.BaseModel{ID: uuid.New()}, Name: "HR"}

		r.hrEmployee.items[uuid.New()] = &model.HREmployee{
			BaseModel: model.BaseModel{ID: uuid.New()},
			FirstName: "A", LastName: "A", HireDate: &recentHire,
			Status: model.EmployeeStatusActive,
		}
		r.hrEmployee.items[uuid.New()] = &model.HREmployee{
			BaseModel: model.BaseModel{ID: uuid.New()},
			FirstName: "B", LastName: "B", HireDate: &oldHire,
			Status: model.EmployeeStatusOnLeave,
		}

		for i := 0; i < 6; i++ {
			id := uuid.New()
			hire := now.AddDate(0, 0, -i)
			r.hrEmployee.items[id] = &model.HREmployee{
				BaseModel: model.BaseModel{ID: id},
				FirstName: "N", LastName: "E", HireDate: &hire,
			}
		}
		for i := 0; i < 5; i++ {
			id := uuid.New()
			r.evaluation.evals[id] = &model.Evaluation{
				BaseModel: model.BaseModel{ID: id, UpdatedAt: now.Add(-time.Duration(i) * time.Hour)},
				Status:    model.EvaluationStatusSubmitted,
			}
			r.announcement.items[uuid.New()] = &model.HRAnnouncement{
				BaseModel: model.BaseModel{ID: uuid.New(), CreatedAt: now.Add(-time.Duration(i) * time.Minute)},
				Title:     "Announcement",
			}
		}

		stats, err := dashboardSvc.GetStats(ctx)
		if err != nil {
			t.Fatalf("GetStats failed: %v", err)
		}
		if stats["new_hires"].(int64) < 1 {
			t.Fatalf("expected at least one new hire")
		}
		if stats["on_leave"].(int64) < 1 {
			t.Fatalf("expected at least one on leave")
		}

		activities, err := dashboardSvc.GetRecentActivities(ctx)
		if err != nil {
			t.Fatalf("GetRecentActivities failed: %v", err)
		}
		if len(activities) != 10 {
			t.Fatalf("expected truncation to 10, got %d", len(activities))
		}

		deps2, _ := setupHRServiceDeps(t)
		emptyActivities, err := NewHRDashboardService(deps2).GetRecentActivities(ctx)
		if err != nil {
			t.Fatalf("GetRecentActivities(empty) failed: %v", err)
		}
		if len(emptyActivities) != 1 || emptyActivities[0]["type"] != "info" {
			t.Fatalf("expected single fallback info activity")
		}
	})

	t.Run("org_chart_get_and_simulate", func(t *testing.T) {
		fromID := uuid.New()
		toID := uuid.New()
		empID := uuid.New()
		r.hrDepartment.items[fromID] = &model.HRDepartment{BaseModel: model.BaseModel{ID: fromID}, Name: "From"}
		r.hrDepartment.items[toID] = &model.HRDepartment{BaseModel: model.BaseModel{ID: toID}, Name: "To"}
		r.hrEmployee.items[empID] = &model.HREmployee{
			BaseModel:    model.BaseModel{ID: empID},
			FirstName:    "Move",
			LastName:     "Target",
			Position:     "Engineer",
			DepartmentID: hrUUIDPtr(fromID),
		}
		chart, err := orgSvc.GetOrgChart(ctx)
		if err != nil || len(chart) < 2 {
			t.Fatalf("GetOrgChart failed: %v", err)
		}

		sim, err := orgSvc.Simulate(ctx, map[string]interface{}{
			"moves": []interface{}{
				map[string]interface{}{
					"employee_id":        empID.String(),
					"from_department_id": fromID.String(),
					"to_department_id":   toID.String(),
				},
				"invalid",
			},
			"renames": []interface{}{
				map[string]interface{}{"department_id": toID.String(), "new_name": "Moved"},
			},
		})
		if err != nil {
			t.Fatalf("Simulate failed: %v", err)
		}
		if len(sim) < 2 {
			t.Fatalf("expected at least two departments after simulate")
		}
		for _, d := range sim {
			if flag, ok := d["simulated"].(bool); !ok || !flag {
				t.Fatalf("every node should have simulated=true")
			}
		}

		r.hrDepartment.findAllErr = errors.New("find all error")
		if _, err := orgSvc.GetOrgChart(ctx); err == nil {
			t.Fatalf("expected GetOrgChart error")
		}
	})
}

func TestHRService_AttendanceIntegration(t *testing.T) {
	ctx := context.Background()

	t.Run("get_integration_branches", func(t *testing.T) {
		deps, r := setupHRServiceDeps(t)
		svc := NewAttendanceIntegrationService(deps)

		r.hrEmployee.useCount = true
		r.hrEmployee.total = 0
		r.hrEmployee.active = 0
		v, err := svc.GetIntegration(ctx, "day", "")
		if err != nil {
			t.Fatalf("GetIntegration failed: %v", err)
		}
		if v["attendance_rate"].(float64) != 0 {
			t.Fatalf("expected attendance_rate=0 when total=0")
		}

		r.hrEmployee.total = 1
		r.hrEmployee.active = 2
		userID := uuid.New()
		r.user.Users[userID] = &model.User{BaseModel: model.BaseModel{ID: userID}, FirstName: "A", LastName: "B"}
		today := time.Now().Truncate(24 * time.Hour)
		for i := 0; i < 2; i++ {
			att := &model.Attendance{
				BaseModel:       model.BaseModel{ID: uuid.New()},
				UserID:          uuid.New(),
				Date:            today,
				Status:          model.AttendanceStatusPresent,
				OvertimeMinutes: 60,
			}
			r.attendance.Attendances[att.ID] = att
		}
		leave := &model.LeaveRequest{
			BaseModel: model.BaseModel{ID: uuid.New()},
			UserID:    userID,
			Status:    model.ApprovalStatusApproved,
			StartDate: today,
			EndDate:   today,
		}
		r.leave.LeaveRequests[leave.ID] = leave

		v, err = svc.GetIntegration(ctx, "day", "")
		if err != nil {
			t.Fatalf("GetIntegration failed: %v", err)
		}
		if v["absent_today"].(int64) != 0 {
			t.Fatalf("expected absent_today to be clamped to 0")
		}
	})

	t.Run("get_alerts_and_trend_branches", func(t *testing.T) {
		deps, r := setupHRServiceDeps(t)
		svc := NewAttendanceIntegrationService(deps)
		r.hrEmployee.useCount = true
		r.hrEmployee.total = 1
		r.hrEmployee.active = 1
		userID := uuid.New()
		r.user.Users[userID] = &model.User{BaseModel: model.BaseModel{ID: userID}, FirstName: "Tarou", LastName: "Yamada"}
		r.overtime.monthlyOvertime[userID] = 46 * 60
		day := time.Now().Truncate(24 * time.Hour)
		r.attendance.Attendances[uuid.New()] = &model.Attendance{BaseModel: model.BaseModel{ID: uuid.New()}, UserID: userID, Date: day}

		alerts, err := svc.GetAlerts(ctx)
		if err != nil {
			t.Fatalf("GetAlerts failed: %v", err)
		}
		if len(alerts) == 0 {
			t.Fatalf("expected overtime alert")
		}

		deps2, r2 := setupHRServiceDeps(t)
		svc2 := NewAttendanceIntegrationService(deps2)
		u2 := uuid.New()
		r2.user.Users[u2] = &model.User{BaseModel: model.BaseModel{ID: u2}, FirstName: "No", LastName: "Attendance"}
		alerts, err = svc2.GetAlerts(ctx)
		if err != nil {
			t.Fatalf("GetAlerts failed: %v", err)
		}
		if len(alerts) == 0 {
			t.Fatalf("expected absence alert")
		}

		for i := 0; i < 30; i++ {
			d := time.Now().AddDate(0, 0, -i).Truncate(24 * time.Hour)
			r.attendance.Attendances[uuid.New()] = &model.Attendance{BaseModel: model.BaseModel{ID: uuid.New()}, Date: d}
			r.attendance.Attendances[uuid.New()] = &model.Attendance{BaseModel: model.BaseModel{ID: uuid.New()}, Date: d}
		}
		trend, err := svc.GetTrend(ctx, "month")
		if err != nil {
			t.Fatalf("GetTrend failed: %v", err)
		}
		if len(trend) != 30 {
			t.Fatalf("expected 30 trend points, got %d", len(trend))
		}
		rate := trend[len(trend)-1]["attendance_rate"].(float64)
		if rate > 100 {
			t.Fatalf("attendance_rate should be capped at 100, got %.1f", rate)
		}
	})
}

func TestHRService_OneOnOneSkillSalary(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()

	oneOnOneSvc := NewOneOnOneService(deps)
	skillSvc := NewSkillService(deps)
	salarySvc := NewSalaryService(deps)

	t.Run("one_on_one_create_update_actions", func(t *testing.T) {
		if _, err := oneOnOneSvc.Create(ctx, model.OneOnOneCreateRequest{
			EmployeeID:    "bad-uuid",
			ScheduledDate: "2024-01-01",
		}, uuid.New()); err == nil {
			t.Fatalf("expected invalid employee_id error")
		}

		empID := uuid.New()
		meeting, err := oneOnOneSvc.Create(ctx, model.OneOnOneCreateRequest{
			EmployeeID:    empID.String(),
			ScheduledDate: "2024-03-01",
			Agenda:        "Agenda",
		}, uuid.New())
		if err != nil {
			t.Fatalf("Create meeting failed: %v", err)
		}
		if meeting.Frequency != "biweekly" {
			t.Fatalf("expected default frequency biweekly")
		}

		mood := "good"
		agenda := "new agenda"
		notes := "notes"
		if _, err := oneOnOneSvc.Update(ctx, meeting.ID, model.OneOnOneUpdateRequest{
			Status: "done", Agenda: &agenda, Notes: &notes, Mood: &mood,
		}); err != nil {
			t.Fatalf("Update meeting failed: %v", err)
		}

		withItem, err := oneOnOneSvc.AddActionItem(ctx, meeting.ID, model.ActionItemRequest{Title: "Task A"})
		if err != nil {
			t.Fatalf("AddActionItem failed: %v", err)
		}
		var items []map[string]interface{}
		_ = jsonUnmarshal(withItem.ActionItems, &items)
		actionID := items[0]["id"].(string)
		toggled, err := oneOnOneSvc.ToggleActionItem(ctx, meeting.ID, actionID)
		if err != nil {
			t.Fatalf("ToggleActionItem failed: %v", err)
		}
		_ = jsonUnmarshal(toggled.ActionItems, &items)
		if items[0]["completed"] != true {
			t.Fatalf("expected action item to be toggled true")
		}

		r.oneOnOne.findByIDErr = errors.New("find error")
		if _, err := oneOnOneSvc.AddActionItem(ctx, meeting.ID, model.ActionItemRequest{Title: "x"}); err == nil {
			t.Fatalf("expected find error")
		}
	})

	t.Run("skill_map_add_update", func(t *testing.T) {
		if _, err := skillSvc.GetSkillMap(ctx, "", "bad-uuid"); err == nil {
			t.Fatalf("expected parse error")
		}
		empID := uuid.New()
		added, err := skillSvc.AddSkill(ctx, empID, model.SkillAddRequest{SkillName: "Go"})
		if err != nil {
			t.Fatalf("AddSkill failed: %v", err)
		}
		if added.Category != "technical" || added.Level != 1 {
			t.Fatalf("default category/level not applied")
		}
		level := 4
		category := "business"
		if _, err := skillSvc.UpdateSkill(ctx, added.ID, model.SkillUpdateRequest{
			Level: &level, Category: &category,
		}); err != nil {
			t.Fatalf("UpdateSkill failed: %v", err)
		}
		if _, err := skillSvc.GetSkillMap(ctx, "", empID.String()); err != nil {
			t.Fatalf("GetSkillMap(employee) failed: %v", err)
		}
		if _, err := skillSvc.GetSkillMap(ctx, "Engineering", ""); err != nil {
			t.Fatalf("GetSkillMap(all) failed: %v", err)
		}
	})

	t.Run("salary_simulate_and_budget", func(t *testing.T) {
		r.hrEmployee.items = map[uuid.UUID]*model.HREmployee{
			uuid.New(): {BaseModel: model.BaseModel{ID: uuid.New()}, Grade: "M1", Position: "Dev", BaseSalary: 500000},
			uuid.New(): {BaseModel: model.BaseModel{ID: uuid.New()}, Grade: "M1", Position: "Dev", BaseSalary: 550000},
		}
		out, err := salarySvc.Simulate(ctx, model.SalarySimulateRequest{
			Grade:           "M1",
			Position:        "Dev",
			YearsOfService:  "5",
			EvaluationScore: "80",
		})
		if err != nil {
			t.Fatalf("Simulate failed: %v", err)
		}
		if out["base_salary"].(float64) <= 0 || out["projected_salary"].(float64) <= 0 {
			t.Fatalf("expected positive salary projection")
		}

		cases := map[string]float64{
			"S1": 250000, "M1": 400000, "M2": 500000, "L1": 600000, "L2": 750000, "X": 350000,
		}
		r.hrEmployee.items = map[uuid.UUID]*model.HREmployee{}
		for grade, expected := range cases {
			got, _ := salarySvc.Simulate(ctx, model.SalarySimulateRequest{Grade: grade})
			if got["base_salary"].(float64) != expected {
				t.Fatalf("grade %s expected base %.0f got %.0f", grade, expected, got["base_salary"].(float64))
			}
		}

		r.salary.overview = map[string]interface{}{"total_payroll": 300000.0, "headcount": int64(0)}
		dept := &model.HRDepartment{BaseModel: model.BaseModel{ID: uuid.New()}, Name: "Dev", Code: "D", Budget: 0}
		r.hrDepartment.items[dept.ID] = dept
		budget, err := salarySvc.GetBudget(ctx, "Dev")
		if err != nil {
			t.Fatalf("GetBudget failed: %v", err)
		}
		if budget["total_budget"].(float64) <= 0 {
			t.Fatalf("expected fallback budget > 0")
		}
		if max(2, 1) != 2 || max(1, 2) != 2 {
			t.Fatalf("max helper branch failed")
		}
	})
}

func TestHRService_OnboardingOffboardingSurvey(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()

	onboardingSvc := NewOnboardingService(deps)
	offboardingSvc := NewOffboardingService(deps)
	surveySvc := NewSurveyService(deps)

	t.Run("onboarding_create_update_toggle_template", func(t *testing.T) {
		empID := uuid.New()
		templateID := uuid.New()
		mentorID := uuid.New()
		o, err := onboardingSvc.Create(ctx, model.OnboardingCreateRequest{
			EmployeeID: empID.String(),
			TemplateID: templateID.String(),
			StartDate:  "2024-04-01",
			MentorID:   mentorID.String(),
		})
		if err != nil {
			t.Fatalf("Create onboarding failed: %v", err)
		}
		if o.TemplateID == nil || o.MentorID == nil {
			t.Fatalf("expected template and mentor to be set")
		}

		if _, err := onboardingSvc.Update(ctx, o.ID, map[string]interface{}{"status": "in_progress"}); err != nil {
			t.Fatalf("Update onboarding failed: %v", err)
		}
		o.Tasks = []byte(`[{"id":"a1","completed":false}]`)
		r.onboarding.items[o.ID] = o
		toggled, err := onboardingSvc.ToggleTask(ctx, o.ID, "a1")
		if err != nil {
			t.Fatalf("ToggleTask failed: %v", err)
		}
		var tasks []map[string]interface{}
		_ = jsonUnmarshal(toggled.Tasks, &tasks)
		if tasks[0]["completed"] != true {
			t.Fatalf("expected task to be toggled true")
		}

		if _, err := onboardingSvc.CreateTemplate(ctx, model.OnboardingTemplateCreateRequest{
			Name: "Default", Tasks: []map[string]interface{}{{"title": "Setup"}},
		}); err != nil {
			t.Fatalf("CreateTemplate failed: %v", err)
		}
	})

	t.Run("offboarding_create_update_toggle_analytics", func(t *testing.T) {
		empID := uuid.New()
		o, err := offboardingSvc.Create(ctx, model.OffboardingCreateRequest{
			EmployeeID:      empID.String(),
			LastWorkingDate: "2024-05-01",
			Reason:          "resignation",
		})
		if err != nil {
			t.Fatalf("Create offboarding failed: %v", err)
		}

		status := "completed"
		note := "done"
		exit := "interview"
		if _, err := offboardingSvc.Update(ctx, o.ID, model.OffboardingUpdateRequest{
			Status: &status, Notes: &note, ExitInterview: &exit,
		}); err != nil {
			t.Fatalf("Update offboarding failed: %v", err)
		}

		o.Checklist = []byte(`[{"key":"pc","completed":false}]`)
		r.offboarding.items[o.ID] = o
		out, err := offboardingSvc.ToggleChecklist(ctx, o.ID, "pc")
		if err != nil {
			t.Fatalf("ToggleChecklist failed: %v", err)
		}
		var list []map[string]interface{}
		_ = jsonUnmarshal(out.Checklist, &list)
		if list[0]["completed"] != true {
			t.Fatalf("expected checklist to be toggled true")
		}

		analytics, err := offboardingSvc.GetAnalytics(ctx)
		if err != nil || analytics == nil {
			t.Fatalf("GetAnalytics failed: %v", err)
		}
	})

	t.Run("survey_create_update_publish_close_results_submit", func(t *testing.T) {
		createdBy := uuid.New()
		s, err := surveySvc.Create(ctx, model.SurveyCreateRequest{
			Title: "Engagement",
		}, createdBy)
		if err != nil {
			t.Fatalf("Create survey failed: %v", err)
		}
		if s.Type != "engagement" {
			t.Fatalf("expected default survey type engagement")
		}

		title := "Updated"
		sType := "pulse"
		isAnonymous := false
		updated, err := surveySvc.Update(ctx, s.ID, model.SurveyUpdateRequest{
			Title:       &title,
			Type:        &sType,
			IsAnonymous: &isAnonymous,
			Questions:   []map[string]interface{}{{"q": "How are you?"}},
		})
		if err != nil {
			t.Fatalf("Update survey failed: %v", err)
		}
		if updated.Title != "Updated" || updated.Type != "pulse" {
			t.Fatalf("survey update not applied")
		}

		if _, err := surveySvc.Publish(ctx, s.ID); err != nil {
			t.Fatalf("Publish failed: %v", err)
		}
		if _, err := surveySvc.Close(ctx, s.ID); err != nil {
			t.Fatalf("Close failed: %v", err)
		}
		empID := uuid.New()
		if err := surveySvc.SubmitResponse(ctx, s.ID, &empID, model.SurveyResponseRequest{
			Answers: []map[string]interface{}{{"q": "1", "a": "5"}},
		}); err != nil {
			t.Fatalf("SubmitResponse failed: %v", err)
		}
		res, err := surveySvc.GetResults(ctx, s.ID)
		if err != nil {
			t.Fatalf("GetResults failed: %v", err)
		}
		if res["total_responses"].(int64) != 1 {
			t.Fatalf("expected total_responses=1")
		}

		r.survey.findByIDErr = errors.New("find error")
		if _, err := surveySvc.Publish(ctx, uuid.New()); err == nil {
			t.Fatalf("expected publish find error")
		}
	})
}

func TestHRService_PassThroughMethods(t *testing.T) {
	deps, r := setupHRServiceDeps(t)
	ctx := context.Background()

	employeeSvc := NewHREmployeeService(deps)
	deptSvc := NewHRDepartmentService(deps)
	evalSvc := NewEvaluationService(deps)
	goalSvc := NewGoalService(deps)
	trainingSvc := NewTrainingService(deps)
	recruitSvc := NewRecruitmentService(deps)
	docSvc := NewDocumentService(deps)
	announceSvc := NewAnnouncementService(deps)
	oneSvc := NewOneOnOneService(deps)
	skillSvc := NewSkillService(deps)
	salarySvc := NewSalaryService(deps)
	onboardingSvc := NewOnboardingService(deps)
	offboardingSvc := NewOffboardingService(deps)
	surveySvc := NewSurveyService(deps)

	empID := uuid.New()
	r.hrEmployee.items[empID] = &model.HREmployee{BaseModel: model.BaseModel{ID: empID}, FirstName: "A", LastName: "B"}
	if _, err := employeeSvc.FindByID(ctx, empID); err != nil {
		t.Fatalf("HREmployee FindByID failed: %v", err)
	}

	deptID := uuid.New()
	r.hrDepartment.items[deptID] = &model.HRDepartment{BaseModel: model.BaseModel{ID: deptID}, Name: "Dept"}
	if _, err := deptSvc.FindByID(ctx, deptID); err != nil {
		t.Fatalf("HRDepartment FindByID failed: %v", err)
	}
	if _, err := deptSvc.FindAll(ctx); err != nil {
		t.Fatalf("HRDepartment FindAll failed: %v", err)
	}
	if err := deptSvc.Delete(ctx, deptID); err != nil {
		t.Fatalf("HRDepartment Delete failed: %v", err)
	}

	cycleID := uuid.New()
	r.evaluation.cycles[cycleID] = &model.EvaluationCycle{BaseModel: model.BaseModel{ID: cycleID}, Name: "Cycle"}
	createdEval, err := evalSvc.Create(ctx, model.EvaluationCreateRequest{
		EmployeeID: empID,
		CycleID:    cycleID,
	}, uuid.New())
	if err != nil {
		t.Fatalf("Evaluation Create failed: %v", err)
	}
	if _, err := evalSvc.FindByID(ctx, createdEval.ID); err != nil {
		t.Fatalf("Evaluation FindByID failed: %v", err)
	}
	if _, _, err := evalSvc.FindAll(ctx, 1, 10, "", ""); err != nil {
		t.Fatalf("Evaluation FindAll failed: %v", err)
	}
	if _, err := evalSvc.FindAllCycles(ctx); err != nil {
		t.Fatalf("Evaluation FindAllCycles failed: %v", err)
	}

	goal, err := goalSvc.Create(ctx, model.HRGoalCreateRequest{Title: "Goal X"}, empID)
	if err != nil {
		t.Fatalf("Goal Create failed: %v", err)
	}
	if _, err := goalSvc.FindByID(ctx, goal.ID); err != nil {
		t.Fatalf("Goal FindByID failed: %v", err)
	}
	if _, _, err := goalSvc.FindAll(ctx, 1, 10, "", "", ""); err != nil {
		t.Fatalf("Goal FindAll failed: %v", err)
	}
	title := "Updated"
	if _, err := goalSvc.Update(ctx, goal.ID, model.HRGoalUpdateRequest{Title: &title}); err != nil {
		t.Fatalf("Goal Update failed: %v", err)
	}
	if err := goalSvc.Delete(ctx, goal.ID); err != nil {
		t.Fatalf("Goal Delete failed: %v", err)
	}

	tr, err := trainingSvc.Create(ctx, model.TrainingProgramCreateRequest{Title: "Train", StartDate: "2024-01-01", EndDate: "2024-01-02"})
	if err != nil {
		t.Fatalf("Training Create failed: %v", err)
	}
	if _, err := trainingSvc.FindByID(ctx, tr.ID); err != nil {
		t.Fatalf("Training FindByID failed: %v", err)
	}
	if _, _, err := trainingSvc.FindAll(ctx, 1, 10, "", ""); err != nil {
		t.Fatalf("Training FindAll failed: %v", err)
	}
	newTitle := "Train 2"
	if _, err := trainingSvc.Update(ctx, tr.ID, model.TrainingProgramUpdateRequest{Title: &newTitle}); err != nil {
		t.Fatalf("Training Update failed: %v", err)
	}
	if err := trainingSvc.Delete(ctx, tr.ID); err != nil {
		t.Fatalf("Training Delete failed: %v", err)
	}

	pos, err := recruitSvc.CreatePosition(ctx, model.PositionCreateRequest{Title: "Backend"})
	if err != nil {
		t.Fatalf("Recruitment CreatePosition failed: %v", err)
	}
	if _, err := recruitSvc.FindPositionByID(ctx, pos.ID); err != nil {
		t.Fatalf("Recruitment FindPositionByID failed: %v", err)
	}
	if _, _, err := recruitSvc.FindAllPositions(ctx, 1, 10, "", ""); err != nil {
		t.Fatalf("Recruitment FindAllPositions failed: %v", err)
	}
	newOpenings := 2
	if _, err := recruitSvc.UpdatePosition(ctx, pos.ID, model.PositionUpdateRequest{Openings: &newOpenings}); err != nil {
		t.Fatalf("Recruitment UpdatePosition failed: %v", err)
	}
	app, err := recruitSvc.CreateApplicant(ctx, model.ApplicantCreateRequest{PositionID: pos.ID, Name: "N", Email: "e@e"})
	if err != nil {
		t.Fatalf("Recruitment CreateApplicant failed: %v", err)
	}
	if _, err := recruitSvc.FindAllApplicants(ctx, app.PositionID.String(), ""); err != nil {
		t.Fatalf("Recruitment FindAllApplicants failed: %v", err)
	}

	doc := &model.HRDocument{Title: "Doc", Type: "pdf", FileName: "a.pdf", FilePath: "/tmp/a.pdf"}
	if err := docSvc.Upload(ctx, doc); err != nil {
		t.Fatalf("Document Upload failed: %v", err)
	}
	if _, _, err := docSvc.FindAll(ctx, 1, 10, "", ""); err != nil {
		t.Fatalf("Document FindAll failed: %v", err)
	}
	if _, err := docSvc.FindByID(ctx, doc.ID); err != nil {
		t.Fatalf("Document FindByID failed: %v", err)
	}
	if err := docSvc.Delete(ctx, doc.ID); err != nil {
		t.Fatalf("Document Delete failed: %v", err)
	}

	ann, err := announceSvc.Create(ctx, model.AnnouncementCreateRequest{Title: "A", Content: "B"}, uuid.New())
	if err != nil {
		t.Fatalf("Announcement Create failed: %v", err)
	}
	if _, err := announceSvc.FindByID(ctx, ann.ID); err != nil {
		t.Fatalf("Announcement FindByID failed: %v", err)
	}
	if _, _, err := announceSvc.FindAll(ctx, 1, 10, ""); err != nil {
		t.Fatalf("Announcement FindAll failed: %v", err)
	}
	if err := announceSvc.Delete(ctx, ann.ID); err != nil {
		t.Fatalf("Announcement Delete failed: %v", err)
	}

	meeting, err := oneSvc.Create(ctx, model.OneOnOneCreateRequest{EmployeeID: empID.String(), ScheduledDate: "2024-01-01"}, uuid.New())
	if err != nil {
		t.Fatalf("OneOnOne Create failed: %v", err)
	}
	if _, err := oneSvc.FindByID(ctx, meeting.ID); err != nil {
		t.Fatalf("OneOnOne FindByID failed: %v", err)
	}
	if _, err := oneSvc.FindAll(ctx, "", ""); err != nil {
		t.Fatalf("OneOnOne FindAll failed: %v", err)
	}
	if err := oneSvc.Delete(ctx, meeting.ID); err != nil {
		t.Fatalf("OneOnOne Delete failed: %v", err)
	}

	r.skill.gapAnalysis = []map[string]interface{}{{"skill_name": "Go", "gap": 1}}
	if _, err := skillSvc.GetGapAnalysis(ctx, "Engineering"); err != nil {
		t.Fatalf("Skill GetGapAnalysis failed: %v", err)
	}

	r.salary.overview = map[string]interface{}{"total_payroll": 1000.0, "headcount": int64(1)}
	if _, err := salarySvc.GetOverview(ctx, ""); err != nil {
		t.Fatalf("Salary GetOverview failed: %v", err)
	}
	r.salary.records[empID] = []model.SalaryRecord{{BaseModel: model.BaseModel{ID: uuid.New()}, EmployeeID: empID}}
	if _, err := salarySvc.GetHistory(ctx, empID); err != nil {
		t.Fatalf("Salary GetHistory failed: %v", err)
	}

	onb, err := onboardingSvc.Create(ctx, model.OnboardingCreateRequest{EmployeeID: empID.String(), StartDate: "2024-01-01"})
	if err != nil {
		t.Fatalf("Onboarding Create failed: %v", err)
	}
	if _, err := onboardingSvc.FindByID(ctx, onb.ID); err != nil {
		t.Fatalf("Onboarding FindByID failed: %v", err)
	}
	if _, err := onboardingSvc.FindAll(ctx, ""); err != nil {
		t.Fatalf("Onboarding FindAll failed: %v", err)
	}
	if _, err := onboardingSvc.CreateTemplate(ctx, model.OnboardingTemplateCreateRequest{Name: "Onb"}); err != nil {
		t.Fatalf("Onboarding CreateTemplate failed: %v", err)
	}
	if _, err := onboardingSvc.FindAllTemplates(ctx); err != nil {
		t.Fatalf("Onboarding FindAllTemplates failed: %v", err)
	}

	off, err := offboardingSvc.Create(ctx, model.OffboardingCreateRequest{
		EmployeeID: empID.String(), LastWorkingDate: "2024-03-31",
	})
	if err != nil {
		t.Fatalf("Offboarding Create failed: %v", err)
	}
	if _, err := offboardingSvc.FindByID(ctx, off.ID); err != nil {
		t.Fatalf("Offboarding FindByID failed: %v", err)
	}
	if _, err := offboardingSvc.FindAll(ctx, ""); err != nil {
		t.Fatalf("Offboarding FindAll failed: %v", err)
	}

	s, err := surveySvc.Create(ctx, model.SurveyCreateRequest{Title: "S"}, uuid.New())
	if err != nil {
		t.Fatalf("Survey Create failed: %v", err)
	}
	if _, err := surveySvc.FindByID(ctx, s.ID); err != nil {
		t.Fatalf("Survey FindByID failed: %v", err)
	}
	if _, err := surveySvc.FindAll(ctx, "", ""); err != nil {
		t.Fatalf("Survey FindAll failed: %v", err)
	}
	if err := surveySvc.Delete(ctx, s.ID); err != nil {
		t.Fatalf("Survey Delete failed: %v", err)
	}
}
