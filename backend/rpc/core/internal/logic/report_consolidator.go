package logic

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"ai_companion/rpc/core/internal/svc"
)

// ConsolidateReport aggregates dialogue evaluations into a single session report.
// It returns: evaluationReport JSON string, overallScore, and error.
// If there are still pending evaluations, it returns a temporary report with overallScore=0, but does NOT write to the DB.
func ConsolidateReport(ctx context.Context, svcCtx *svc.ServiceContext, sessionId string) (string, int32, error) {
	// 1. Query the session
	session, err := svcCtx.PracticeSessionModel.FindOne(ctx, sessionId)
	if err != nil {
		return "", 0, fmt.Errorf("failed to find session: %v", err)
	}

	// 2. Query dialogues
	type dbDialogue struct {
		Role       string         `db:"role"`
		Content    string         `db:"content"`
		Evaluation sql.NullString `db:"evaluation"`
	}
	var dialogues []dbDialogue
	queryDialogues := "SELECT role, content, evaluation FROM dialogues WHERE practice_session_id = ? ORDER BY created_at ASC, role ASC"
	err = svcCtx.SqlConn.QueryRowsCtx(ctx, &dialogues, queryDialogues, sessionId)
	if err != nil {
		return "", 0, fmt.Errorf("failed to query dialogues: %v", err)
	}

	// 3. Check if there are dialogues and if any user dialogue is still evaluating
	hasDialogues := false
	hasEvaluating := false
	var totalScore float64
	var count int
	var fluencyList []string
	var relevanceList []string
	var logicList []string
	var depthList []string
	var starList []string
	var overallCommentList []string

	for _, d := range dialogues {
		if d.Role == "user" {
			if d.Content != "" {
				hasDialogues = true
			}
			if d.Evaluation.Valid {
				if d.Evaluation.String == "evaluating" {
					hasEvaluating = true
				} else if d.Evaluation.String != "" {
					var round map[string]interface{}
					cleanJSON := strings.ReplaceAll(d.Evaluation.String, "```json", "")
					cleanJSON = strings.ReplaceAll(cleanJSON, "```", "")
					cleanJSON = strings.TrimSpace(cleanJSON)
					if err := json.Unmarshal([]byte(cleanJSON), &round); err == nil {
						// Extract score (fallback to average of scores sub-object if root score is missing)
						var s float64
						if val, ok := round["score"]; ok {
							if fVal, ok := val.(float64); ok {
								s = fVal
							}
						} else if val, ok := round["overall_score"]; ok {
							if fVal, ok := val.(float64); ok {
								s = fVal
							}
						} else if scoresVal, ok := round["scores"].(map[string]interface{}); ok {
							var sum float64
							var subCount float64
							for _, subName := range []string{"fluency", "vocabulary", "grammar", "pronunciation"} {
								if sv, ok := scoresVal[subName].(float64); ok {
									sum += sv
									subCount++
								}
							}
							if subCount > 0 {
								s = sum / subCount
							}
						}
						if s > 0 {
							totalScore += s
							count++
						}

						// Extract dimensions
						if dims, ok := round["dimensions"].(map[string]interface{}); ok {
							if val, ok := dims["fluency"].(string); ok && val != "" {
								fluencyList = append(fluencyList, val)
							}
							if val, ok := dims["relevance"].(string); ok && val != "" {
								relevanceList = append(relevanceList, val)
							}
							if val, ok := dims["logic"].(string); ok && val != "" {
								logicList = append(logicList, val)
							}
							if val, ok := dims["depth"].(string); ok && val != "" {
								depthList = append(depthList, val)
							}
							if val, ok := dims["star_alignment"].(string); ok && val != "" {
								starList = append(starList, val)
							}
						}

						// Extract overall comment
						if comment, ok := round["overall_comment"].(string); ok && comment != "" {
							overallCommentList = append(overallCommentList, comment)
						}
					}
				}
			}
		}
	}

	var evaluationReport string
	var overallScore int32

	if count > 0 {
		overallScore = int32(totalScore / float64(count))
		joinFeedback := func(list []string, defaultMsg string) string {
			if len(list) == 0 {
				return defaultMsg
			}
			
			var cleanedList []string
			for _, item := range list {
				cleaned := cleanParagraph(item)
				if cleaned != "" {
					cleanedList = append(cleanedList, cleaned)
				}
			}
			
			if len(cleanedList) == 0 {
				return defaultMsg
			}
			
			// Limit to the last 3 turns to keep the commentary concise and not too long
			if len(cleanedList) > 3 {
				cleanedList = cleanedList[len(cleanedList)-3:]
			}
			
			if len(cleanedList) == 1 {
				return cleanedList[0]
			}
			return strings.Join(cleanedList, " ")
		}

		reportObj := map[string]interface{}{
			"overall_score": overallScore,
			"dimensions": map[string]string{
				"fluency":   joinFeedback(fluencyList, "评估未见异常。"),
				"relevance": joinFeedback(relevanceList, "评估未见异常。"),
				"logic":     joinFeedback(logicList, "评估未见异常。"),
				"depth":     joinFeedback(depthList, "评估未见异常。"),
			},
			"overall_comment": joinFeedback(overallCommentList, "本次口语练习已完成，AI 口语评估分析成功。您在各维度的发音、流畅度以及表达内容均已获得多维度深度反馈，请参考下方的详细指标与重写建议进行针对性优化提升。"),
		}

		reportBytes, _ := json.Marshal(reportObj)
		evaluationReport = string(reportBytes)

		// 🌟 只有在会话已结束，并且没有任何对话处于 "evaluating" 状态时，才持久化写入到 practice_sessions 的 evaluation_report 缓存字段中
		if session.Status == "completed" && !hasEvaluating {
			session.EvaluationReport = sql.NullString{String: evaluationReport, Valid: true}
			_ = svcCtx.PracticeSessionModel.Update(ctx, session)
		}
	} else {
		var pendingMsg string
		if hasEvaluating || hasDialogues {
			pendingMsg = "AI 正在分析您的回答，请稍后刷新查看完整评估报告。"
		} else {
			pendingMsg = "本次练习未进行充分的对话，无法生成评估报告。"
		}

		reportObj := map[string]interface{}{
			"overall_score": 0,
			"dimensions": map[string]string{
				"fluency":   pendingMsg,
				"relevance": pendingMsg,
				"logic":     pendingMsg,
				"depth":     pendingMsg,
			},
			"overall_comment": pendingMsg,
		}
		reportBytes, _ := json.Marshal(reportObj)
		evaluationReport = string(reportBytes)
		overallScore = 0
	}

	return evaluationReport, overallScore, nil
}

func cleanParagraph(text string) string {
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\t", " ")

	// Remove markers at the very beginning of the string
	reStart := regexp.MustCompile(`^\s*([-[*•]|\d+[\.)]|\(\d+\))\s*`)
	text = reStart.ReplaceAllString(text, "")

	// Remove markers in the middle of the string
	reMiddle := regexp.MustCompile(`\s*([-[*•]|\d+[\.)]|\(\d+\))\s+`)
	text = reMiddle.ReplaceAllString(text, " ")

	// Normalize spacing
	words := strings.Fields(text)
	text = strings.Join(words, " ")

	return text
}
