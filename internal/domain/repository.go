package domain

import (
	"context"

	"github.com/philiaspace/phi-core/domain"
	examd "github.com/philiaspace/phi-exam-domain/domain"
)

// QuestionRepository defines the contract for question bank persistence.
type QuestionRepository interface {
	domain.Repository[Question]
	FindByLevelAndSection(ctx context.Context, level examd.JLPTLevel, section examd.Section, limit int) ([]Question, error)
	CountByLevelAndSection(ctx context.Context, level examd.JLPTLevel, section examd.Section) (int, error)
	SearchQuestions(ctx context.Context, level examd.JLPTLevel, section examd.Section, search string, sort string, sortDir string, limit int, offset int) ([]Question, int, error)
	FindByPassageID(ctx context.Context, id examd.PassageID) ([]Question, error)
	FindBySourceGroupKey(ctx context.Context, key string) ([]Question, error)
	FindWithOptions(ctx context.Context, id examd.QuestionID) (*Question, []Option, error)
	FindByIDs(ctx context.Context, ids []examd.QuestionID) ([]Question, error)
	FindPassageByID(ctx context.Context, id examd.PassageID) (*Passage, error)
	ListPassages(ctx context.Context, level examd.JLPTLevel, section examd.Section, limit int) ([]Passage, error)
	FindAssetsForQuestions(ctx context.Context, ids []examd.QuestionID) (map[examd.QuestionID][]Asset, error)
	FindAssetsForPassages(ctx context.Context, ids []examd.PassageID) (map[examd.PassageID][]Asset, error)
	FindAssetByID(ctx context.Context, id string) (*Asset, error)
	FindAssetsByQuestionID(ctx context.Context, questionID string) ([]Asset, error)
	ListAssets(ctx context.Context, assetType string, limit int, offset int) ([]Asset, int, error)
	ListTemplates(ctx context.Context, level examd.JLPTLevel) ([]PackageTemplate, error)
	SaveOptions(ctx context.Context, questionID string, options []Option) error
}
