package usecase_test

import (
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/infrastructure/storage/pg"
	boardrepo "github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/boards/repository"
	boardpg "github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/boards/repository/boardpg"
	postrepo "github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/posts/repository"
	postpg "github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/posts/repository/postpg"
	threadrepo "github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/threads/repository"
	threadpg "github.com/ShudrikNikolai/gl1tch/backend/boards/internal/modules/threads/repository/threadpg"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/transport/http_api"
	"github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase"
	boarduc "github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/boards"
	postuc "github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/posts"
	threaduc "github.com/ShudrikNikolai/gl1tch/backend/boards/internal/usecase/threads"
)

// Проверка на этапе компиляции: реальные репозитории и pg.Client подходят под интерфейсы usecase.
// Если сигнатура разойдётся, сборка (а не рантайм) укажет, где именно.
var (
	_ boarduc.Repo = (*boardpg.Repo)(nil)

	_ threaduc.BoardRepo  = (*boardpg.Repo)(nil)
	_ threaduc.ThreadRepo = (*threadpg.Repo)(nil)
	_ threaduc.PostRepo   = (*postpg.Repo)(nil)
	_ threaduc.Tx         = (*pg.Client)(nil)

	_ postuc.BoardRepo  = (*boardpg.Repo)(nil)
	_ postuc.ThreadRepo = (*threadpg.Repo)(nil)
	_ postuc.PostRepo   = (*postpg.Repo)(nil)
	_ postuc.Tx         = (*pg.Client)(nil)

	_ usecase.BoardUseCase  = (*boarduc.UseCase)(nil)
	_ usecase.ThreadUseCase = (*threaduc.UseCase)(nil)
	_ usecase.PostUseCase   = (*postuc.UseCase)(nil)

	_ httpapi.Pinger = (*pg.Client)(nil)
)

// В di репозитории хранятся в полях-интерфейсах модулей: они тоже должны подходить под usecase.
var (
	_ boarduc.Repo = boardrepo.Repository(nil)

	_ threaduc.BoardRepo  = boardrepo.Repository(nil)
	_ threaduc.ThreadRepo = threadrepo.Repository(nil)
	_ threaduc.PostRepo   = postrepo.Repository(nil)

	_ postuc.BoardRepo  = boardrepo.Repository(nil)
	_ postuc.ThreadRepo = threadrepo.Repository(nil)
	_ postuc.PostRepo   = postrepo.Repository(nil)
)
