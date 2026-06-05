import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { GlassCard } from '../components/GlassCard';
import { api } from '../services/api';
import { CommentaryScorecard } from '../components/CommentaryScorecard';
import { ChevronLeft, FileText, Award, Cpu } from 'lucide-react';
import './ReportPage.css';


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
               {data.transcript.split('\n').map((line: string, i: number) => {
                 if (!line.trim()) return null;
                 const isAI = line.startsWith('AI:');
                 return (
                   <div key={i} className={`transcript-line ${isAI ? 'ai' : 'user'}`}>
                     <span className="line-prefix">{isAI ? 'AI TEACHER' : 'YOU'}</span>
                    <p>{line.replace(/^(AI:|User:)/, '')}</p>
                  </div>
                );
              })}
            </div>
          </GlassCard>

          <div className="report-sidebar">
            <GlassCard className="recommendation-card">
              <div className="card-header">
                <Cpu className="accent-icon" />
                <h2>AI Recommendations</h2>
              </div>
              <div className="rec-content">
                <p>Based on your logic score, we suggest focusing more on the <strong>STAR framework</strong> for behavioral questions.</p>
                <div className="rec-tag">Recommended: High-Level Logic Training</div>
              </div>
            </GlassCard>
          </div>
        </div>
      </div>
    </div>
  );
};

