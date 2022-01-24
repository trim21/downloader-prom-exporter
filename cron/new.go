package cron

import (
	"github.com/hekmon/transmissionrpc/v2"
	"github.com/robfig/cron/v3"

	"app/pkg/errgo"
	"app/pkg/logger"
)

type C struct {
	c *cron.Cron
}

func (c *C) Run() {
	c.c.Run()
}

func New(tr *transmissionrpc.Client) (*C, error) {
	logger.Info("creating cron manager")
	c := cron.New()

	if err := setupTransmissionMetrics(tr, c); err != nil {
		return nil, errgo.Wrap(err, "cron")
	}

	return &C{c}, nil
}
