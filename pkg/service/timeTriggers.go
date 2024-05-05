package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"git.edtech.vm.prod-6.cloud.el/fabric/api-client"
	"git.edtech.vm.prod-6.cloud.el/fabric/lib"
	"git.edtech.vm.prod-6.cloud.el/fabric/models"
	"git.edtech.vm.prod-6.cloud.el/packages/logger"
	duration "github.com/xhit/go-str2duration"
	"go.uber.org/zap"

	app_lib "git.edtech.vm.prod-6.cloud.el/fabric/app/lib"
)

const queryTimeTriggers = "query_sys_triggers_whits_active_timer"

func timeTriggers(ctx context.Context, api api.Api, interval time.Duration) {
	t := time.NewTicker(interval)

	for {
		select {
		case <-ctx.Done():
			t.Stop()

			return

		case <-t.C:
			res, err := api.Query(ctx, queryTimeTriggers, "GET", "")
			if err != nil {
				logger.Error(ctx, "unable get time triggers", zap.Error(err))

				continue
			}

			var triggers models.ResponseData
			err = json.Unmarshal([]byte(res), &triggers)
			if err != nil {
				logger.Error(ctx, "failed deserialize time triggers", zap.Error(err))

				continue
			}

			for _, trigger := range triggers.Data {
				timeTriggerProcess(ctx, api, trigger)
			}
		}
	}
}

func timeTriggerProcess(ctx context.Context, api api.Api, trigger models.Data) {
	startDate, _ := trigger.Attr("datestart", "value")
	startTime, _ := trigger.Attr("timestart", "value")
	endDate, _ := trigger.Attr("datestop", "value")
	endTime, _ := trigger.Attr("timestop", "value")
	start := time.Now()
	now := start.UTC().Truncate(time.Minute)
	ctx = logger.SetFieldCtx(ctx, "trigger", trigger.Uid)

	if startDate != "" {
		if startTime == "" {
			startTime = "00:00"
		}

		parsed := app_lib.Funcs.Timeparseany(fmt.Sprintf("%s %s:00 MSK", startDate, startTime), true)
		if parsed.Err != "" {
			logger.Error(ctx, "unable parse trigger start time", zap.String("start-date", startDate),
				zap.String("start-time", startTime), zap.String("error", parsed.Err))

			return
		}

		if now.Before(parsed.Time) {
			return
		}
	}

	if endDate != "" {
		if endTime == "" {
			endTime = "23:59"
		}

		parsed := app_lib.Funcs.Timeparseany(fmt.Sprintf("%s %s:00 MSK", endDate, endTime), true)
		if parsed.Err != "" {
			logger.Error(ctx, "unable parse trigger end time", zap.String("end-date", endDate),
				zap.String("end-time", endTime), zap.String("error", parsed.Err))

			return
		}

		if now.After(parsed.Time) {
			api.ObjAttrUpdate(ctx, trigger.Uid, "timer", "", "", "")

			return
		}
	}

	interval, _ := trigger.Attr("interval", "value")
	if !checkMatchInterval(now, interval) {
		return
	}

	url, _ := trigger.Attr("query_url", "value")
	method, _ := trigger.Attr("query_type", "src")
	body, _ := trigger.Attr("query_body", "value")

	if method == "" {
		method = "GET"
	}

	if url == "" {
		logger.Error(ctx, "empty url")

		return
	}

	_, err := lib.Curl(ctx, method, url, body, nil, nil, nil)
	if err != nil {
		logger.Error(ctx, "unable send trigger request", zap.Float64("timing", time.Since(start).Seconds()), zap.Error(err))
	} else {
		logger.Info(ctx, "trigger completed", zap.Float64("timing", time.Since(start).Seconds()))
	}
}

func checkMatchInterval(t time.Time, interval string) bool {
	parsed, err := duration.Str2Duration(strings.ReplaceAll(interval, ":", "h") + "m")
	if err != nil {
		return false
	}

	return t.Equal(t.Truncate(parsed))
}
