import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { GlassCard } from '../components/GlassCard';
import { api } from '../services/api';
import { CommentaryScorecard } from '../components/CommentaryScorecard';
import { ChevronLeft, FileText, Award } from 'lucide-react';
import './ReportPage.css';

const getAudioPlayUrl = (url: string) => {
  if (!url) return '';
  try {
    const urlObj = new URL(url);
    return `/minio${urlObj.pathname}${urlObj.search}`;
  } catch (e) {
    return url;
  }
};

const RoundScoreBar: React.FC<{ label: string; value: number; color: string }> = ({ label, value, color }) => {
  const [width, setWidth] = useState(0);

  useEffect(() => {
    // Triggers smooth width animation when the bar mounts/expands
    const timer = setTimeout(() => setWidth(value), 50);
    return () => clearTimeout(timer);
  }, [value]);

  return (
    <div className="round-score-bar-item animate-fade-in">
      <div className="round-score-bar-info">
        <span className="round-score-bar-label">{label}</span>
        <span className="round-score-bar-value">{value}%</span>
      </div>
      <div className="round-score-bar-track">
        <div className="round-score-bar-fill" style={{ width: `${width}%`, background: color }} />
      </div>
    </div>
  );
};

const RoundEvaluationBlock: React.FC<{ evaluation: any }> = ({ evaluation }) => {
  const [expanded, setExpanded] = useState(true);
  const dims = evaluation.dimensions || {};
  const comment = evaluation.overall_comment || "";
  const scores = evaluation.scores || {};
  const hasScores = scores.fluency !== undefined || scores.vocabulary !== undefined || scores.grammar !== undefined || scores.pronunciation !== undefined;

  return (
    <div className="round-evaluation-wrapper">
      <button 
        onClick={() => setExpanded(!expanded)} 
        className={`toggle-eval-btn ${expanded ? 'active' : ''}`}
      >
        <span>{expanded ? "Hide Round Performance" : "Show Round Performance"}</span>
        <span className="toggle-icon">{expanded ? "▲" : "▼"}</span>
      </button>
      
      {expanded && (
        <div className="round-evaluation-details animate-slide-down">
          {hasScores && (
            <div className="round-scores-container">
              <span className="round-scores-title">Speaking Competencies</span>
              <div className="round-scores-grid">
                <RoundScoreBar label="Fluency & Flow" value={scores.fluency || 0} color="linear-gradient(90deg, #3b82f6, #60a5fa)" />
                <RoundScoreBar label="Vocabulary & Word" value={scores.vocabulary || 0} color="linear-gradient(90deg, #6366f1, #818cf8)" />
                <RoundScoreBar label="Grammar & Accuracy" value={scores.grammar || 0} color="linear-gradient(90deg, #10b981, #34d399)" />
                <RoundScoreBar label="Pronunciation" value={scores.pronunciation || 0} color="linear-gradient(90deg, #f97316, #fb923c)" />
              </div>
            </div>
          )}
          {comment && (
            <div className="eval-detail-comment">
              <strong>Round Comment:</strong> {comment}
            </div>
          )}
          <div className="eval-detail-dims">
            {Object.entries(dims).map(([key, val]) => {
              if (!val) return null;
              const labels: Record<string, string> = {
                fluency: "Fluency & Flow Feedback",
                relevance: "Vocabulary & Word Choice Feedback",
                logic: "Grammar & Accuracy Feedback",
                depth: "Refined Rewrite Feedback",
                star_alignment: "Speaking Tips"
              };
              return (
                <div key={key} className="eval-dim-row">
                  <span className="dim-label">{labels[key] || key}</span>
                  <span className="dim-val">{String(val)}</span>
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
};

export const ReportPage: React.FC = () => {
  const { sessionId } = useParams<{ sessionId: string }>();
  const navigate = useNavigate();
  const [data, setData] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const res = await api.getSessionDetail(sessionId!);
        setData(res);
      } catch (err) {
        console.error('Failed to fetch report:', err);
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, [sessionId]);

  if (loading) return <div className="report-loading">Loading neural analysis...</div>;
  if (!data) return <div className="report-error">Simulation data not found.</div>;

  const parsedCommentary = data.commentary ? JSON.parse(data.commentary.replace(/```json|```/g, '')) : null;

  // Parse dialogues from transcript (support both structured JSON array and old plain text fallback)
  let dialogues: Array<{ role: 'ai' | 'user'; content: string; audio_url?: string; evaluation?: any }> = [];
  try {
    dialogues = JSON.parse(data.transcript);
  } catch (e) {
    dialogues = data.transcript.split('\n')
      .filter((line: string) => line.trim())
      .map((line: string) => {
        const isAI = line.startsWith('AI:');
        return {
          role: isAI ? 'ai' : 'user',
          content: line.replace(/^(AI:|User:)/, '')
        };
      });
  }

  return (
    <div className="report-container animate-fade-in">
      <div className="report-nav">
        <button onClick={() => navigate('/')} className="back-btn">
          <ChevronLeft size={20} /> Back to Hub
        </button>
        <h1>Practice Summary: {data.scenario_name}</h1>
      </div>
 
      <div className="report-grid">
        <div className="report-summary-header">
          <GlassCard className="score-summary-card">
            <div className="card-header">
              <Award className="accent-icon" />
              <h2>Performance Matrix</h2>
            </div>
            {parsedCommentary ? (
              <CommentaryScorecard data={parsedCommentary} />
            ) : (
              <div className="no-score">Analysis not finalized.</div>
            )}
          </GlassCard>
        </div>
 
        <div className="report-details-grid">
          <GlassCard className="transcript-card">
            <div className="card-header">
              <FileText className="accent-icon" />
              <h2>Neural Transcript</h2>
            </div>
            <div className="transcript-content">
              {dialogues.map((item, idx) => {
                const isAI = item.role === 'ai';
                let score = 0;
                let hasEval = false;
                if (item.evaluation) {
                  if (item.evaluation.score !== undefined || item.evaluation.overall_score !== undefined) {
                    score = item.evaluation.score ?? item.evaluation.overall_score;
                    hasEval = true;
                  } else if (item.evaluation.scores) {
                    const s = item.evaluation.scores;
                    const sum = (s.fluency || 0) + (s.vocabulary || 0) + (s.grammar || 0) + (s.pronunciation || 0);
                    score = Math.round(sum / 4);
                    hasEval = score > 0;
                  }
                }
                
                return (
                  <div key={idx} className={`transcript-line-container ${isAI ? 'ai' : 'user'}`}>
                    <div className={`transcript-bubble ${isAI ? 'ai' : 'user'}`}>
                      <div className="bubble-header">
                        <span className="line-prefix">{isAI ? 'AI TEACHER' : 'YOU'}</span>
                        {hasEval && (
                          <span className={`round-score-badge ${score >= 80 ? 'high' : score >= 60 ? 'mid' : 'low'}`}>
                            Score: {score}
                          </span>
                        )}
                      </div>
                      <p className="bubble-text">{item.content}</p>
                      
                      {/* Audio player for user voice reply */}
                      {!isAI && item.audio_url && (
                        <div className="audio-player-container">
                          <audio 
                            src={getAudioPlayUrl(item.audio_url)} 
                            controls 
                            className="user-audio-player"
                          />
                        </div>
                      )}
 
                      {/* Display round evaluation if present */}
                      {item.evaluation && (
                        <RoundEvaluationBlock evaluation={item.evaluation} />
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          </GlassCard>
        </div>
      </div>
    </div>
  );
};
