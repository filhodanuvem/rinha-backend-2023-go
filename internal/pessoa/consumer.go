package pessoa

import (
	"log/slog"
	"time"

	"github.com/filhodanuvem/rinha"
	"github.com/filhodanuvem/rinha/internal/config"
)

func Consume(chPessoas chan rinha.Pessoa, chExit chan struct{}, repo *Repository, batch int) {
	slog.Debug("Starting consumer...")
	defer slog.Debug("Finishing consumer...")
	i := 0
	pessoas := make([]rinha.Pessoa, 0, batch)
	tick := time.NewTicker(config.WorkerTimeout)

	for {
		select {
		case p, ok := <-chPessoas:
			pessoas = append(pessoas, p)
			i++
			if i == batch || !ok {
				if err := repo.Insert(pessoas); err != nil {
					slog.Error(err.Error())
				}
				i = 0
				pessoas = make([]rinha.Pessoa, 0, batch)
			}

			if !ok {
				chExit <- struct{}{}
				return
			}

		case <-tick.C:
			if len(pessoas) > 0 {
				if err := repo.Insert(pessoas); err != nil {
					slog.Error(err.Error())
				}
			}
			i = 0
			pessoas = make([]rinha.Pessoa, 0, batch)
		}
	}
}
