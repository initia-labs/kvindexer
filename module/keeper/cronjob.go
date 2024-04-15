package keeper

func (k Keeper) RegisterCronJob(cronjob Cronjob) error {
	return k.crontab.RegisterJobWithPattern(cronjob.Pattern, cronjob.Tag, cronjob.Initialize, cronjob.Job)
}

// just one-liner to register cronjob
func (k Keeper) RegisterCronjobs(cronjobs ...Cronjob) error {
	for _, cronjob := range cronjobs {
		err := k.crontab.RegisterJobWithPattern(cronjob.Pattern, cronjob.Tag, cronjob.Initialize, cronjob.Job)
		if err != nil {
			return err
		}
	}
	return nil
}

type Cronjob struct {
	// Tag must be unique
	Tag string
	// Pattern is a cron pattern
	Pattern string
	// Initializer is a function that will be called when the cronjob is started
	Initialize JobInitializer
	// Job is a function that will be called when the cronjob is running
	Job JobFunc
}
