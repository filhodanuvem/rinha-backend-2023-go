package pessoa

import (
	"log/slog"
	"time"

	"github.com/filhodanuvem/rinha"
)

func Consume(chPessoas chan rinha.Pessoa, chExit chan struct{}, repo *Repository, limit int) {
	slog.Info("Starting consumer...")
	defer slog.Info("Finishing consumer...")
	i := 0
	pessoas := make([]rinha.Pessoa, 0, limit)
	tick := time.NewTicker(3 * time.Second)

	for {
		select {
		case p, ok := <-chPessoas:
			pessoas = append(pessoas, p)
			i++
			if i == limit || !ok {
				if err := repo.Insert(pessoas); err != nil {
					slog.Error(err.Error())
				}
				i = 0
				pessoas = make([]rinha.Pessoa, 0, limit)
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
			pessoas = make([]rinha.Pessoa, 0, limit)
		}
	}
}
