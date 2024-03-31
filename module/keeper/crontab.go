package keeper

import (
	"fmt"
	"reflect"
	"runtime"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/initia-labs/indexer/v2/config"
	"github.com/pkg/errors"
)

type JobInitializer func(keeper *Keeper, config config.CronjobConfig) error
type JobFunc func(keeper *Keeper, config config.CronjobConfig) error

type Crontab struct {
	config      *config.IndexerConfig
	keeper      *Keeper
	scheduler   *gocron.Scheduler
	initializer map[string]JobInitializer
}

func NewCrontab(c *config.IndexerConfig, keeper *Keeper) *Crontab {
	tz, _ := time.LoadLocation("Local")
	ct := &Crontab{
		config:      c,
		keeper:      keeper,
		scheduler:   gocron.NewScheduler(tz),
		initializer: map[string]JobInitializer{},
	}

	return ct
}

func (ct *Crontab) Initialize() error {
	for tag, initializer := range ct.initializer {
		err := initializer(ct.keeper, ct.config.CronjobConfigs[tag])
		if err != nil {
			return errors.Wrap(err, "failed to initialize cron")
		}
	}
	return nil
}

func (ct *Crontab) Start() {
	ct.scheduler.StartAsync()
}

func (ct *Crontab) Stop() {
	ct.scheduler.Stop()
}

func (ct *Crontab) RegisterJobWithPattern(pattern, tag string, jobInit JobInitializer, jobFunc JobFunc) error {
	// originally tag can be duplicated but we don't allow it and we uses it as a unique key
	_, err := ct.scheduler.FindJobsByTag(tag)
	if err == nil {
		panic(fmt.Errorf("%+v already exists", jobFunc))
	}
	sched, err := ct.scheduler.Cron(pattern).Do(func() {
		err := jobFunc(ct.keeper, ct.config.CronjobConfigs[tag])
		if err != nil {
			panic(errors.Wrap(err, "failed to run cron"))
		}
	})
	if err != nil {
		panic(errors.Wrap(err, "failed to register cron"))
	}
	sched.Tag(tag)
	ct.initializer[tag] = jobInit
	return nil
}

func (ct *Crontab) RegisterJob(tag string, jobInit JobInitializer, jobFunc JobFunc) error {
	//initName := getFunctionName(jobInit)
	//jobName := getFunctionName(jobFunc)
	if !ct.config.IsEnabledCronjob(tag) {
		return nil
	}
	pattern, ok := ct.config.CronjobConfigs[tag]["pattern"].(string)
	if !ok {
		panic(fmt.Errorf("pattern is not string"))
	}
	return ct.RegisterJobWithPattern(pattern, tag, jobInit, jobFunc)
}

func (ct *Crontab) UnregisterJob(tag string) {
	_ = ct.scheduler.RemoveByTag(tag)
}

func (ct *Crontab) IsRegisteredJob(tag string) bool {
	_, err := ct.scheduler.FindJobsByTag(tag)
	return err == nil
}

//nolint:unused
func getFunctionName(temp interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(temp).Pointer()).Name()
}
