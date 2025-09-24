package projection

import _model "bitbucket.org/edts/go-task-management/internal/model"

type TeamSummary struct {
	Team        *_model.Team
	MemberCount *int32
}
