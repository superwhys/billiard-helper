package dto

type FeedbackReportReq struct {
	Operator
	Content string `json:"content"`
}
