// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

//go:build !migration
// +build !migration

package migration

import (
	"fmt"

	"github.com/alimy/tryst/cfg"
	"github.com/sirupsen/logrus"
)

func Run() error {
	if cfg.If("Migration") {
		return fmt.Errorf("Migration feature requested, but this binary was built without the migration tag")
	}
	logrus.Infoln("skip migrate because Migration feature is disabled")
	return nil
}
