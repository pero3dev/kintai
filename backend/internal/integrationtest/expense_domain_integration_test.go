package integrationtest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/your-org/kintai/backend/internal/model"
)

type expenseListResponse struct {
	Data       []model.Expense `json:"data"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

type expenseMonthlyTrendResponse struct {
	Data []model.MonthlyTrendItem `json:"data"`
}

type expenseCommentsResponse struct {
	Data []model.ExpenseCommentResponse `json:"data"`
}

type expenseHistoriesResponse struct {
	Data []model.ExpenseHistoryResponse `json:"data"`
}

type expenseTemplatesResponse struct {
	Data []model.ExpenseTemplate `json:"data"`
}

type expensePoliciesResponse struct {
	Data []model.ExpensePolicy `json:"data"`
}

type expenseViolationsResponse struct {
	Data []model.ExpensePolicyViolation `json:"data"`
}

type expenseNotificationsResponse struct {
	Data []model.ExpenseNotification `json:"data"`
}

type expenseRemindersResponse struct {
	Data []model.ExpenseReminder `json:"data"`
}

type expenseDelegatesResponse struct {
	Data []model.ExpenseDelegate `json:"data"`
}

type expenseUseTemplateResponse struct {
	ID uuid.UUID `json:"id"`
}

func TestExpenseDomainIntegration(t *testing.T) {
	env := NewTestEnv(t, nil)
	require.NoError(t, model.ExpenseAutoMigrate(env.DB))
	t.Cleanup(func() {
		_ = os.RemoveAll("uploads")
	})

	t.Run("expense create with items category and total amount calculation", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		employee, employeeHeaders := createActorWithHeaders(t, env, model.RoleEmployee)

		createResp := env.DoJSON(t, http.MethodPost, "/api/v1/expenses", map[string]any{
			"title":  "Business Trip",
			"status": model.ExpenseStatusDraft,
			"notes":  "trip notes",
			"items": []map[string]any{
				{
					"expense_date": "2031-01-10",
					"category":     model.ExpenseCategoryTransportation,
					"description":  "train",
					"amount":       1200.0,
					"receipt_url":  "/uploads/r1.png",
				},
				{
					"expense_date": "2031-01-10",
					"category":     model.ExpenseCategoryMeals,
					"description":  "lunch",
					"amount":       800.0,
					"receipt_url":  "/uploads/r2.png",
				},
			},
		}, employeeHeaders)
		require.Equal(t, http.StatusCreated, createResp.Code)

		var created model.Expense
		require.NoError(t, json.Unmarshal(createResp.Body.Bytes(), &created))
		require.Equal(t, employee.ID, created.UserID)
		require.Equal(t, 2000.0, created.TotalAmount)
		require.Equal(t, model.ExpenseStatusDraft, created.Status)
		require.Len(t, created.Items, 2)
		require.Equal(t, model.ExpenseCategoryTransportation, created.Items[0].Category)
		require.Equal(t, model.ExpenseCategoryMeals, created.Items[1].Category)

		var historyCount int64
		require.NoError(t, env.DB.Model(&model.ExpenseHistory{}).Where("expense_id = ?", created.ID).Count(&historyCount).Error)
		require.Equal(t, int64(1), historyCount)
	})

	t.Run("expense update with status change and delete", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		_, employeeHeaders := createActorWithHeaders(t, env, model.RoleEmployee)
		expense := mustCreateExpense(t, env, employeeHeaders, model.ExpenseStatusDraft)

		updateResp := env.DoJSON(t, http.MethodPut, "/api/v1/expenses/"+expense.ID.String(), map[string]any{
			"title":  "Updated Expense",
			"status": model.ExpenseStatusPending,
			"notes":  "updated",
			"items": []map[string]any{
				{
					"expense_date": "2031-01-15",
					"category":     model.ExpenseCategorySupplies,
					"description":  "pens",
					"amount":       300.0,
					"receipt_url":  "/uploads/r3.png",
				},
				{
					"expense_date": "2031-01-15",
					"category":     model.ExpenseCategoryCommunication,
					"description":  "sim",
					"amount":       700.0,
					"receipt_url":  "/uploads/r4.png",
				},
			},
		}, employeeHeaders)
		require.Equal(t, http.StatusOK, updateResp.Code)

		var updated model.Expense
		require.NoError(t, json.Unmarshal(updateResp.Body.Bytes(), &updated))
		require.Equal(t, "Updated Expense", updated.Title)
		require.Equal(t, model.ExpenseStatusPending, updated.Status)
		require.Equal(t, 1000.0, updated.TotalAmount)

		deleteResp := env.DoJSON(t, http.MethodDelete, "/api/v1/expenses/"+expense.ID.String(), nil, employeeHeaders)
		require.Equal(t, http.StatusOK, deleteResp.Code)

		getResp := env.DoJSON(t, http.MethodGet, "/api/v1/expenses/"+expense.ID.String(), nil, employeeHeaders)
		require.Equal(t, http.StatusNotFound, getResp.Code)
	})

	t.Run("approve and reject with status transition and notification", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		employee, employeeHeaders := createActorWithHeaders(t, env, model.RoleEmployee)
		approver, approverHeaders := createActorWithHeaders(t, env, model.RoleManager)

		pendingForApprove := mustCreateExpense(t, env, employeeHeaders, model.ExpenseStatusPending)

		approveResp := env.DoJSON(t, http.MethodPut, "/api/v1/expenses/"+pendingForApprove.ID.String()+"/approve", map[string]any{
			"status": model.ExpenseStatusApproved,
		}, approverHeaders)
		require.Equal(t, http.StatusOK, approveResp.Code)

		var approved model.Expense
		require.NoError(t, env.DB.First(&approved, "id = ?", pendingForApprove.ID).Error)
		require.Equal(t, model.ExpenseStatusApproved, approved.Status)
		require.NotNil(t, approved.ApprovedBy)
		require.Equal(t, approver.ID, *approved.ApprovedBy)

		var approvedNotif model.ExpenseNotification
		require.NoError(
			t,
			env.DB.Where("user_id = ? AND expense_id = ? AND type = ?", employee.ID, pendingForApprove.ID, "approved").
				First(&approvedNotif).Error,
		)

		pendingForReject := mustCreateExpense(t, env, employeeHeaders, model.ExpenseStatusPending)
		rejectResp := env.DoJSON(t, http.MethodPut, "/api/v1/expenses/"+pendingForReject.ID.String()+"/approve", map[string]any{
			"status":          model.ExpenseStatusRejected,
			"rejected_reason": "policy violation",
		}, approverHeaders)
		require.Equal(t, http.StatusOK, rejectResp.Code)

		var rejected model.Expense
		require.NoError(t, env.DB.First(&rejected, "id = ?", pendingForReject.ID).Error)
		require.Equal(t, model.ExpenseStatusRejected, rejected.Status)
		require.Equal(t, "policy violation", rejected.RejectedReason)

		var rejectedNotif model.ExpenseNotification
		require.NoError(
			t,
			env.DB.Where("user_id = ? AND expense_id = ? AND type = ?", employee.ID, pendingForReject.ID, "rejected").
				First(&rejectedNotif).Error,
		)
	})

	t.Run("list search report and monthly aggregation", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		_, employeeHeaders := createActorWithHeaders(t, env, model.RoleEmployee)
		e1 := mustCreateExpense(t, env, employeeHeaders, model.ExpenseStatusPending)
		e2 := mustCreateExpense(t, env, employeeHeaders, model.ExpenseStatusApproved)

		var first, second model.Expense
		require.NoError(t, env.DB.Preload("Items").First(&first, "id = ?", e1.ID).Error)
		require.NoError(t, env.DB.Preload("Items").First(&second, "id = ?", e2.ID).Error)

		require.NoError(t, env.DB.Model(&first).Update("created_at", mustDateTime(t, "2031-02-10T10:00:00")).Error)
		require.NoError(t, env.DB.Model(&second).Update("created_at", mustDateTime(t, "2031-03-10T10:00:00")).Error)
		require.NoError(t, env.DB.Model(&first.Items[0]).Update("category", model.ExpenseCategoryTransportation).Error)
		require.NoError(t, env.DB.Model(&second.Items[0]).Update("category", model.ExpenseCategoryMeals).Error)

		listStatusResp := env.DoJSON(
			t,
			http.MethodGet,
			"/api/v1/expenses?status=pending&page=1&page_size=20",
			nil,
			employeeHeaders,
		)
		require.Equal(t, http.StatusOK, listStatusResp.Code)
		var listByStatus expenseListResponse
		require.NoError(t, json.Unmarshal(listStatusResp.Body.Bytes(), &listByStatus))
		require.Equal(t, int64(1), listByStatus.Total)
		require.Equal(t, model.ExpenseStatusPending, listByStatus.Data[0].Status)

		listCategoryResp := env.DoJSON(
			t,
			http.MethodGet,
			"/api/v1/expenses?category=transportation&page=1&page_size=20",
			nil,
			employeeHeaders,
		)
		require.Equal(t, http.StatusOK, listCategoryResp.Code)
		var listByCategory expenseListResponse
		require.NoError(t, json.Unmarshal(listCategoryResp.Body.Bytes(), &listByCategory))
		require.Equal(t, int64(1), listByCategory.Total)
		require.Equal(t, e1.ID, listByCategory.Data[0].ID)

		pendingResp := env.DoJSON(t, http.MethodGet, "/api/v1/expenses/pending?page=1&page_size=20", nil, employeeHeaders)
		require.Equal(t, http.StatusOK, pendingResp.Code)
		var pendingList expenseListResponse
		require.NoError(t, json.Unmarshal(pendingResp.Body.Bytes(), &pendingList))
		require.Equal(t, int64(1), pendingList.Total)

		reportResp := env.DoJSON(
			t,
			http.MethodGet,
			"/api/v1/expenses/report?start_date=2031-02-01&end_date=2031-03-31",
			nil,
			employeeHeaders,
		)
		require.Equal(t, http.StatusOK, reportResp.Code)
		var report model.ExpenseReportResponse
		require.NoError(t, json.Unmarshal(reportResp.Body.Bytes(), &report))
		require.Greater(t, report.TotalAmount, 0.0)
		require.NotEmpty(t, report.CategoryBreakdown)

		monthlyResp := env.DoJSON(
			t,
			http.MethodGet,
			"/api/v1/expenses/report/monthly?year=2031",
			nil,
			employeeHeaders,
		)
		require.Equal(t, http.StatusOK, monthlyResp.Code)
		var monthly expenseMonthlyTrendResponse
		require.NoError(t, json.Unmarshal(monthlyResp.Body.Bytes(), &monthly))
		require.NotEmpty(t, monthly.Data)
	})

	t.Run("csv and pdf export", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		_, headers := createActorWithHeaders(t, env, model.RoleEmployee)
		mustCreateExpense(t, env, headers, model.ExpenseStatusApproved)

		csvBody, csvHeaders, csvCode := env.DoDownload(
			t,
			http.MethodGet,
			"/api/v1/expenses/export/csv?start_date=2031-01-01&end_date=2031-12-31",
			headers,
		)
		require.Equal(t, http.StatusOK, csvCode)
		require.Contains(t, csvHeaders.Get("Content-Type"), "text/csv")
		require.Contains(t, csvHeaders.Get("Content-Disposition"), "expenses.csv")
		require.GreaterOrEqual(t, len(csvBody), 3)
		require.Equal(t, byte(0xEF), csvBody[0])
		require.Equal(t, byte(0xBB), csvBody[1])
		require.Equal(t, byte(0xBF), csvBody[2])

		pdfBody, pdfHeaders, pdfCode := env.DoDownload(
			t,
			http.MethodGet,
			"/api/v1/expenses/export/pdf?start_date=2031-01-01&end_date=2031-12-31",
			headers,
		)
		require.Equal(t, http.StatusOK, pdfCode)
		require.Contains(t, pdfHeaders.Get("Content-Type"), "application/pdf")
		require.Contains(t, pdfHeaders.Get("Content-Disposition"), "expenses.pdf")
		require.Greater(t, len(pdfBody), 8)
		require.Equal(t, "%PDF-1.4\n", string(pdfBody[:9]))
	})

	t.Run("receipt upload multipart", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		_, headers := createActorWithHeaders(t, env, model.RoleEmployee)
		resp := env.DoMultipart(
			t,
			http.MethodPost,
			"/api/v1/expenses/receipts/upload",
			nil,
			map[string]MultipartFile{
				"file": {
					FileName: "receipt.txt",
					Content:  []byte("receipt-content"),
				},
			},
			headers,
		)
		require.Equal(t, http.StatusOK, resp.Code)

		var upload model.ReceiptUploadResponse
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &upload))
		require.True(
			t,
			len(upload.URL) > 0 &&
				(strings.Contains(upload.URL, "/uploads/receipts/") || strings.Contains(upload.URL, "/uploads\\receipts\\")),
		)
	})

	t.Run("comments add and get and history fetch", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		employee, employeeHeaders := createActorWithHeaders(t, env, model.RoleEmployee)
		other, otherHeaders := createActorWithHeaders(t, env, model.RoleManager)
		expense := mustCreateExpense(t, env, employeeHeaders, model.ExpenseStatusPending)

		addResp := env.DoJSON(
			t,
			http.MethodPost,
			"/api/v1/expenses/"+expense.ID.String()+"/comments",
			map[string]any{"content": "please review"},
			otherHeaders,
		)
		require.Equal(t, http.StatusCreated, addResp.Code)
		var comment model.ExpenseCommentResponse
		require.NoError(t, json.Unmarshal(addResp.Body.Bytes(), &comment))
		require.Equal(t, expense.ID, comment.ExpenseID)
		require.Equal(t, other.ID, comment.UserID)
		require.Equal(t, "please review", comment.Content)

		getCommentsResp := env.DoJSON(
			t,
			http.MethodGet,
			"/api/v1/expenses/"+expense.ID.String()+"/comments",
			nil,
			employeeHeaders,
		)
		require.Equal(t, http.StatusOK, getCommentsResp.Code)
		var comments expenseCommentsResponse
		require.NoError(t, json.Unmarshal(getCommentsResp.Body.Bytes(), &comments))
		require.Len(t, comments.Data, 1)
		require.Equal(t, comment.ID, comments.Data[0].ID)

		updateResp := env.DoJSON(
			t,
			http.MethodPut,
			"/api/v1/expenses/"+expense.ID.String(),
			map[string]any{
				"title":  "History Test Updated",
				"status": model.ExpenseStatusPending,
			},
			employeeHeaders,
		)
		require.Equal(t, http.StatusOK, updateResp.Code)

		historyResp := env.DoJSON(
			t,
			http.MethodGet,
			"/api/v1/expenses/"+expense.ID.String()+"/history",
			nil,
			employeeHeaders,
		)
		require.Equal(t, http.StatusOK, historyResp.Code)
		var histories expenseHistoriesResponse
		require.NoError(t, json.Unmarshal(historyResp.Body.Bytes(), &histories))
		require.GreaterOrEqual(t, len(histories.Data), 2)

		var notif model.ExpenseNotification
		require.NoError(
			t,
			env.DB.Where("user_id = ? AND expense_id = ? AND type = ?", employee.ID, expense.ID, "comment").
				First(&notif).Error,
		)
	})

	t.Run("template CRUD and apply template", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		_, headers := createActorWithHeaders(t, env, model.RoleEmployee)

		createResp := env.DoJSON(t, http.MethodPost, "/api/v1/expenses/templates", map[string]any{
			"name":          "Monthly Pass",
			"title":         "Commute",
			"category":      model.ExpenseCategoryTransportation,
			"description":   "monthly train pass",
			"amount":        10000.0,
			"is_recurring":  true,
			"recurring_day": 1,
		}, headers)
		require.Equal(t, http.StatusCreated, createResp.Code)
		var created model.ExpenseTemplate
		require.NoError(t, json.Unmarshal(createResp.Body.Bytes(), &created))

		updateResp := env.DoJSON(
			t,
			http.MethodPut,
			"/api/v1/expenses/templates/"+created.ID.String(),
			map[string]any{
				"name":     "Monthly Pass Updated",
				"title":    "Commute Updated",
				"category": model.ExpenseCategoryTransportation,
				"amount":   12000.0,
			},
			headers,
		)
		require.Equal(t, http.StatusOK, updateResp.Code)

		getResp := env.DoJSON(t, http.MethodGet, "/api/v1/expenses/templates", nil, headers)
		require.Equal(t, http.StatusOK, getResp.Code)
		var templates expenseTemplatesResponse
		require.NoError(t, json.Unmarshal(getResp.Body.Bytes(), &templates))
		require.Len(t, templates.Data, 1)
		require.Equal(t, created.ID, templates.Data[0].ID)

		useResp := env.DoJSON(
			t,
			http.MethodPost,
			"/api/v1/expenses/templates/"+created.ID.String()+"/use",
			nil,
			headers,
		)
		require.Equal(t, http.StatusCreated, useResp.Code)
		var useResult expenseUseTemplateResponse
		require.NoError(t, json.Unmarshal(useResp.Body.Bytes(), &useResult))

		var used model.Expense
		require.NoError(t, env.DB.Preload("Items").First(&used, "id = ?", useResult.ID).Error)
		require.Equal(t, 12000.0, used.TotalAmount)
		require.Len(t, used.Items, 1)

		deleteResp := env.DoJSON(
			t,
			http.MethodDelete,
			"/api/v1/expenses/templates/"+created.ID.String(),
			nil,
			headers,
		)
		require.Equal(t, http.StatusOK, deleteResp.Code)
	})

	t.Run("policy CRUD and policy violation alerts", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		employee, headers := createActorWithHeaders(t, env, model.RoleEmployee)

		createResp := env.DoJSON(t, http.MethodPost, "/api/v1/expenses/policies", map[string]any{
			"category":               model.ExpenseCategoryMeals,
			"monthly_limit":          5000.0,
			"per_claim_limit":        1000.0,
			"auto_approve_limit":     300.0,
			"requires_receipt_above": 500.0,
			"description":            "meal policy",
			"is_active":              true,
		}, headers)
		require.Equal(t, http.StatusCreated, createResp.Code)
		var created model.ExpensePolicy
		require.NoError(t, json.Unmarshal(createResp.Body.Bytes(), &created))

		updateResp := env.DoJSON(
			t,
			http.MethodPut,
			"/api/v1/expenses/policies/"+created.ID.String(),
			map[string]any{
				"category":               model.ExpenseCategoryMeals,
				"monthly_limit":          6000.0,
				"per_claim_limit":        1200.0,
				"auto_approve_limit":     400.0,
				"requires_receipt_above": 700.0,
				"description":            "meal policy updated",
				"is_active":              true,
			},
			headers,
		)
		require.Equal(t, http.StatusOK, updateResp.Code)

		getPoliciesResp := env.DoJSON(t, http.MethodGet, "/api/v1/expenses/policies", nil, headers)
		require.Equal(t, http.StatusOK, getPoliciesResp.Code)
		var policies expensePoliciesResponse
		require.NoError(t, json.Unmarshal(getPoliciesResp.Body.Bytes(), &policies))
		require.Len(t, policies.Data, 1)

		expense := mustCreateExpense(t, env, headers, model.ExpenseStatusPending)
		violation := model.ExpensePolicyViolation{
			ExpenseID: expense.ID,
			PolicyID:  created.ID,
			UserID:    employee.ID,
			Reason:    "over limit",
			Severity:  "warning",
		}
		require.NoError(t, env.DB.Create(&violation).Error)

		violationsResp := env.DoJSON(t, http.MethodGet, "/api/v1/expenses/policy-violations", nil, headers)
		require.Equal(t, http.StatusOK, violationsResp.Code)
		var violations expenseViolationsResponse
		require.NoError(t, json.Unmarshal(violationsResp.Body.Bytes(), &violations))
		require.Len(t, violations.Data, 1)
		require.Equal(t, violation.ID, violations.Data[0].ID)

		deleteResp := env.DoJSON(
			t,
			http.MethodDelete,
			"/api/v1/expenses/policies/"+created.ID.String(),
			nil,
			headers,
		)
		require.Equal(t, http.StatusOK, deleteResp.Code)
	})

	t.Run("notification settings reminder and notification history", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		employee, headers := createActorWithHeaders(t, env, model.RoleEmployee)

		expense := mustCreateExpense(t, env, headers, model.ExpenseStatusPending)

		notif1 := model.ExpenseNotification{
			UserID:    employee.ID,
			ExpenseID: &expense.ID,
			Type:      "approved",
			Message:   "approved notification",
			IsRead:    false,
		}
		notif2 := model.ExpenseNotification{
			UserID:    employee.ID,
			ExpenseID: &expense.ID,
			Type:      "comment",
			Message:   "comment notification",
			IsRead:    true,
		}
		require.NoError(t, env.DB.Create(&notif1).Error)
		require.NoError(t, env.DB.Create(&notif2).Error)

		due := mustDate(t, "2031-04-01")
		reminder := model.ExpenseReminder{
			UserID:      employee.ID,
			ExpenseID:   &expense.ID,
			Message:     "submit expense",
			DueDate:     &due,
			IsDismissed: false,
		}
		require.NoError(t, env.DB.Create(&reminder).Error)

		notificationsResp := env.DoJSON(t, http.MethodGet, "/api/v1/expenses/notifications?filter=unread", nil, headers)
		require.Equal(t, http.StatusOK, notificationsResp.Code)
		var notifications expenseNotificationsResponse
		require.NoError(t, json.Unmarshal(notificationsResp.Body.Bytes(), &notifications))
		require.Len(t, notifications.Data, 1)
		require.Equal(t, notif1.ID, notifications.Data[0].ID)

		markReadResp := env.DoJSON(
			t,
			http.MethodPut,
			"/api/v1/expenses/notifications/"+notif1.ID.String()+"/read",
			nil,
			headers,
		)
		require.Equal(t, http.StatusOK, markReadResp.Code)

		markAllResp := env.DoJSON(t, http.MethodPut, "/api/v1/expenses/notifications/read-all", nil, headers)
		require.Equal(t, http.StatusOK, markAllResp.Code)

		settingsResp := env.DoJSON(t, http.MethodGet, "/api/v1/expenses/notification-settings", nil, headers)
		require.Equal(t, http.StatusOK, settingsResp.Code)
		var settings model.ExpenseNotificationSetting
		require.NoError(t, json.Unmarshal(settingsResp.Body.Bytes(), &settings))
		require.Equal(t, employee.ID, settings.UserID)

		updateSettingsResp := env.DoJSON(
			t,
			http.MethodPut,
			"/api/v1/expenses/notification-settings",
			map[string]any{
				"email_enabled":        false,
				"push_enabled":         true,
				"approval_alerts":      false,
				"reimbursement_alerts": true,
				"policy_alerts":        false,
				"weekly_digest":        true,
			},
			headers,
		)
		require.Equal(t, http.StatusOK, updateSettingsResp.Code)

		remindersResp := env.DoJSON(t, http.MethodGet, "/api/v1/expenses/reminders", nil, headers)
		require.Equal(t, http.StatusOK, remindersResp.Code)
		var reminders expenseRemindersResponse
		require.NoError(t, json.Unmarshal(remindersResp.Body.Bytes(), &reminders))
		require.Len(t, reminders.Data, 1)
		require.Equal(t, reminder.ID, reminders.Data[0].ID)

		dismissResp := env.DoJSON(
			t,
			http.MethodPut,
			"/api/v1/expenses/reminders/"+reminder.ID.String()+"/dismiss",
			nil,
			headers,
		)
		require.Equal(t, http.StatusOK, dismissResp.Code)
	})

	t.Run("expense approval flow and delegate create approve and CRUD", func(t *testing.T) {
		require.NoError(t, env.ResetDB())

		employee, employeeHeaders := createActorWithHeaders(t, env, model.RoleEmployee)
		delegate, _ := createActorWithHeaders(t, env, model.RoleManager)
		_, approverHeaders := createActorWithHeaders(t, env, model.RoleAdmin)

		flow := model.ExpenseApprovalFlow{
			Name:             "Expense Flow",
			MinAmount:        0,
			MaxAmount:        1000000,
			RequiredSteps:    2,
			IsActive:         true,
			AutoApproveBelow: 1000,
		}
		require.NoError(t, env.DB.Create(&flow).Error)

		getConfigResp := env.DoJSON(t, http.MethodGet, "/api/v1/expenses/approval-flow", nil, employeeHeaders)
		require.Equal(t, http.StatusOK, getConfigResp.Code)
		var config model.ExpenseApprovalFlow
		require.NoError(t, json.Unmarshal(getConfigResp.Body.Bytes(), &config))
		require.Equal(t, flow.ID, config.ID)

		setDelegateResp := env.DoJSON(t, http.MethodPost, "/api/v1/expenses/delegates", map[string]any{
			"delegate_to": delegate.ID.String(),
			"start_date":  "2031-05-01",
			"end_date":    "2031-05-31",
		}, employeeHeaders)
		require.Equal(t, http.StatusCreated, setDelegateResp.Code)
		var createdDelegate model.ExpenseDelegate
		require.NoError(t, json.Unmarshal(setDelegateResp.Body.Bytes(), &createdDelegate))
		require.Equal(t, employee.ID, createdDelegate.UserID)
		require.Equal(t, delegate.ID, createdDelegate.DelegateID)

		getDelegatesResp := env.DoJSON(t, http.MethodGet, "/api/v1/expenses/delegates", nil, employeeHeaders)
		require.Equal(t, http.StatusOK, getDelegatesResp.Code)
		var delegates expenseDelegatesResponse
		require.NoError(t, json.Unmarshal(getDelegatesResp.Body.Bytes(), &delegates))
		require.Len(t, delegates.Data, 1)
		require.Equal(t, createdDelegate.ID, delegates.Data[0].ID)

		expense := mustCreateExpense(t, env, employeeHeaders, model.ExpenseStatusPending)

		step1Resp := env.DoJSON(
			t,
			http.MethodPut,
			"/api/v1/expenses/"+expense.ID.String()+"/advanced-approve",
			map[string]any{
				"action": "approve",
				"step":   1,
			},
			approverHeaders,
		)
		require.Equal(t, http.StatusOK, step1Resp.Code)

		var afterStep1 model.Expense
		require.NoError(t, env.DB.First(&afterStep1, "id = ?", expense.ID).Error)
		require.Equal(t, model.ExpenseStatusStep1Approved, afterStep1.Status)

		returnResp := env.DoJSON(
			t,
			http.MethodPut,
			"/api/v1/expenses/"+expense.ID.String()+"/advanced-approve",
			map[string]any{
				"action": "return",
				"reason": "need fix",
			},
			approverHeaders,
		)
		require.Equal(t, http.StatusOK, returnResp.Code)

		var afterReturn model.Expense
		require.NoError(t, env.DB.First(&afterReturn, "id = ?", expense.ID).Error)
		require.Equal(t, model.ExpenseStatusReturned, afterReturn.Status)

		removeDelegateResp := env.DoJSON(
			t,
			http.MethodDelete,
			"/api/v1/expenses/delegates/"+createdDelegate.ID.String(),
			nil,
			employeeHeaders,
		)
		require.Equal(t, http.StatusOK, removeDelegateResp.Code)
	})
}

func mustCreateExpense(
	t *testing.T,
	env *TestEnv,
	headers map[string]string,
	status model.ExpenseStatus,
) *model.Expense {
	t.Helper()

	title := fmt.Sprintf("it-expense-%s", uuid.NewString())
	resp := env.DoJSON(t, http.MethodPost, "/api/v1/expenses", map[string]any{
		"title":  title,
		"status": status,
		"notes":  "integration",
		"items": []map[string]any{
			{
				"expense_date": "2031-01-01",
				"category":     model.ExpenseCategorySupplies,
				"description":  "item",
				"amount":       1500.0,
				"receipt_url":  "/uploads/r.png",
			},
		},
	}, headers)
	require.Equal(t, http.StatusCreated, resp.Code)

	var expense model.Expense
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &expense))
	return &expense
}

func mustDateTime(t *testing.T, raw string) time.Time {
	t.Helper()

	parsed, err := time.Parse("2006-01-02T15:04:05", raw)
	require.NoError(t, err)
	return parsed
}
