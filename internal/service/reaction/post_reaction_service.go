package reactionsvc

import domain "01.tomorrow-school.ai/git/amadiuly/forum/internal/domain/reaction"

type PostReactionService struct {
	postReactionRepository domain.PostReactionRepository
}

func NewPostReactionService() *PostReactionService {
	return
}