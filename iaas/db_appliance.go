package iaas

import (
	"errors"
	"fmt"
	"net/http"

	"context"
	log "github.com/Sirupsen/logrus"
	"github.com/sacloud/libsacloud/api"
	"github.com/sacloud/libsacloud/sacloud"
	"github.com/sacloud/open-service-broker-sacloud/service/params"
	"github.com/sacloud/open-service-broker-sacloud/util/random"
	"time"
)

type dbApplianceClient struct {
	*client
	createParamFunc func() *sacloud.CreateDatabaseValue
}

func (c *dbApplianceClient) Read(instanceID string) (*sacloud.Database, error) {
	client := c.getRawClient()
	results, err := client.Database.Reset().WithNameLike(instanceID).Find()
	if err != nil {
		return nil, err
	}
	if len(results.Databases) == 0 {
		return nil, api.NewError(http.StatusNotFound, &sacloud.ResultErrorValue{})
	}

	if len(results.Databases) > 1 {
		return nil, errors.New("Multiple resources with the same instance ID is exists")
	}

	return &results.Databases[0], nil
}

func (c *dbApplianceClient) Create(instanceID string, param *params.DatabaseCreateParameter) (*sacloud.Database, error) {

	client := c.getRawClient()

	p := c.createParamFunc()
	p.Plan = sacloud.DatabasePlan(param.PlanID)
	p.SwitchID = fmt.Sprintf("%d", param.SwitchID)
	p.DefaultUser = param.Username
	if p.DefaultUser == "" {
		p.DefaultUser = random.String(10)
	}
	p.UserPassword = random.String(20)

	p.IPAddress1 = param.IPAddress
	p.MaskLen = int(param.MaskLen)
	p.DefaultRoute = param.DefaultRoute

	p.ServicePort = fmt.Sprintf("%d", param.Port)
	p.SourceNetwork = param.AllowNetworks

	p.Tags = []string{markerTag}

	p.Name = instanceID
	createArgs := sacloud.CreateNewDatabase(p)

	return client.Database.Create(createArgs)
}

func (c *dbApplianceClient) Delete(instanceID string) error {

	logFields := log.Fields{
		"instanceID": instanceID,
	}
	log.WithFields(logFields).Debug("IaaS delete instance start")

	db, err := c.Read(instanceID)
	if err != nil {
		return err
	}

	go c.delete(instanceID, db.ID)
	return nil
}

func (c *dbApplianceClient) delete(instanceID string, id int64) {
	logFields := log.Fields{
		"instanceID": instanceID,
	}

	strID := fmt.Sprintf("%d", id)
	mutex.Lock(strID)
	defer mutex.Unlock(strID)

	client := c.getRawClient()

	db, err := client.Database.Read(id)
	if err != nil {
		if e, ok := err.(api.Error); ok && e.ResponseCode() == http.StatusNotFound {
			return
		}

		logFields["err"] = err
		log.WithFields(logFields).Error(
			`IaaS delete instance error: Reading database is failed`)
		return
	}

	if db.IsMigrating() {
		err = c.waitUntilRunning(instanceID, id)
		if err != nil {
			return
		}
	}

	c.stopAndDelete(instanceID, db)
}

func (c *dbApplianceClient) waitUntilRunning(instanceID string, id int64) error {
	logFields := log.Fields{
		"instanceID": instanceID,
	}
	client := c.rawClient

	err := client.Database.SleepUntilUp(id, client.DefaultTimeoutDuration)
	if err != nil {
		logFields["err"] = err
		log.WithFields(logFields).Error(
			`IaaS delete instance error: migrate wait timed out`)
		return err
	}

	// wait for running
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	completeChan := make(chan bool)
	go func() {
		for {
			var res *sacloud.DatabaseStatus
			res, err = client.Database.Status(id)
			if err != nil {
				if e, ok := err.(api.Error); ok && e.ResponseCode() == http.StatusNotFound {
					continue
				}
			}
			if res.Status == "running" {
				completeChan <- true
				return
			}
			time.Sleep(5 * time.Second)
		}
	}()

	select {
	case <-completeChan:
		// noop
	case <-ctx.Done():
		logFields["err"] = ctx.Err()
		log.WithFields(logFields).Error(
			`IaaS delete instance error: startup wait timed out`)
	}

	return nil
}

func (c *dbApplianceClient) stopAndDelete(instanceID string, db *sacloud.Database) {
	logFields := log.Fields{
		"instanceID": instanceID,
	}
	var err error
	client := c.getRawClient()

	// refresh
	db, err = client.Database.Read(db.ID)
	if err != nil {
		if _, ok := err.(api.Error); ok {
			return
		}
		logFields["err"] = err
		log.WithFields(logFields).Error(
			`IaaS delete instance error: Reading database is failed`)
		return
	}

	if db.IsUp() || db.IsFailed() {
		if db.IsUp() {
			_, err = client.Database.Stop(db.ID)
			if err != nil {
				logFields["err"] = err
				log.WithFields(logFields).Error(
					`IaaS delete instance error: error stopping database`)
				return
			}

			// wait for shutdown
			err = client.Database.SleepUntilDown(db.ID, client.DefaultTimeoutDuration)
			if err != nil {
				logFields["err"] = err
				log.WithFields(logFields).Error(
					`IaaS delete instance error: shutdown wait timed out`)
				return
			}
		}
		// delete
		_, err = client.Database.Delete(db.ID)
		if err != nil {
			logFields["err"] = err
			log.WithFields(logFields).Error(
				`IaaS delete instance error: database Delete API is failed`)
			return
		}
	}

}
