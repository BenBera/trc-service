package crontask

import (
	"bitbucket.org/maybets/kra-service/app/constants"
	"context"
	"database/sql"
	goutils "github.com/mudphilo/go-utils"
	"github.com/sirupsen/logrus"
	"time"
)

func (cron *Crontask) SendDashboardReports(ctx context.Context) {

	ctx, span := cron.Tracer.Start(ctx, "SendDashboardReports")
	defer span.End()

	ticker := time.NewTicker(1 * time.Minute)

	for _ = range ticker.C {

		//cron.sendDashboardBetReports(ctx)
		//cron.sendDashboardBetSlipReports(ctx)
		//cron.sendDashboardWinningReports(ctx)
		//cron.sendDashboardBetStatusReports(ctx)
		//cron.sendDashboardLeaderboardReports(ctx)

	}

	select {}
}

func getFields(data map[string]interface{}, excludes []string) []string {

	var fields []string

	for k := range data {

		if !goutils.Contains(excludes, k) {

			fields = append(fields, k)
		}
	}

	return fields
}

func (cron *Crontask) getLastTimeStamp(ctx context.Context, reportName string) string {

	ctx, span := cron.Tracer.Start(ctx, "getLastTimeStamp")
	defer span.End()

	dbUtils := goutils.Db{DB: cron.DB, Context: ctx}
	dbUtils.SetQuery("SELECT last_timestamp FROM reports_sync WHERE report_name = ? ")
	dbUtils.SetParams(reportName)

	var lastTimestamp sql.NullTime

	err := dbUtils.FetchOneWithContext().Scan(&lastTimestamp)
	if err == sql.ErrNoRows {

		return ""
	}

	if err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "error retrieving last_timestamp  ",
				constants.DATA:        reportName,
			}).
			Error(err.Error())

		return ""
	}

	return goutils.ToMysql(lastTimestamp.Time)

}

func (cron *Crontask) setLastTimeStamp(ctx context.Context, reportName, lastTimestamp string) error {

	ctx, span := cron.Tracer.Start(ctx, "setLastTimeStamp")
	defer span.End()

	dbUtils := goutils.Db{DB: cron.DB, Context: ctx}

	inserts := map[string]interface{}{
		"report_name":    reportName,
		"last_timestamp": lastTimestamp,
	}

	_, err := dbUtils.UpsertWithContext("reports_sync", inserts, []string{"last_timestamp"})
	if err != nil {

		logrus.WithContext(ctx).
			WithFields(logrus.Fields{
				constants.DESCRIPTION: "error inserting reports_sync  ",
				constants.DATA:        reportName,
			}).
			Error(err.Error())

	}

	return err

}
