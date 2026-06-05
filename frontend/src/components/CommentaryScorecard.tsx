import React from 'react';
import './CommentaryScorecard.css';

interface CommentaryScorecardProps {
  data: {
    score: number;
    dimensions: {
      fluency: string;
      relevance: string;
      logic: string;
      depth: string;
      star_alignment: string;
    };
    overall_comment: string;
  };
}

export const CommentaryScorecard: React.FC<CommentaryScorecardProps> = ({ data }) => {
  return (
    <div className="scorecard-container animate-slide-up">
      <div className="score-header">
        <div className="score-circle">
          <svg viewBox="0 0 36 36" className="circular-chart">
            <path className="circle-bg"
              d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
            />
            <path className="circle"
              strokeDasharray={`${data.score}, 100`}
              d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
            />
            <text x="18" y="20.35" className="percentage">{data.score}</text>
          </svg>
        </div>
        <div className="overall-summary">
          <h3>Overall Assessment</h3>
          <p>{data.overall_comment}</p>
        </div>
      </div>

      <div className="dimensions-grid">
        <DimensionItem label="Fluency & Flow" value={data.dimensions.fluency} />
        <DimensionItem label="Vocabulary & Word Choice" value={data.dimensions.relevance} />
        <DimensionItem label="Grammar & Accuracy" value={data.dimensions.logic} />
        <DimensionItem label="Refined Rewrite" value={data.dimensions.depth} />
        {data.dimensions.star_alignment && (
          <DimensionItem label="Speaking Tips" value={data.dimensions.star_alignment} />
        )}
      </div>
    </div>
  );
};

const DimensionItem: React.FC<{ label: string, value: string }> = ({ label, value }) => (
  <div className="dimension-item">
    <span className="dimension-label">{label}</span>
    <p className="dimension-value">{value}</p>
  </div>
);
