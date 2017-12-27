package backup

import (
	"fmt"
	"log"

	"github.com/pkg/errors"
	"gopkg.in/mgo.v2/bson"

	mgo "gopkg.in/mgo.v2"
)

// https://docs.mongodb.com/manual/reference/command/

func BalancerStart(sess *mgo.Session) error {
	var result string
	err := sess.Run(bson.M{"balancerStart": 1}, &result)
	if err != nil {
		log.Fatalf("unable to start the mongos balancer: %v\n", err)
	}

	log.Printf(result)
	return nil
}

func BalancerStop(sess *mgo.Session) error {
	var result string
	err := sess.Run(bson.M{"balancerStop": 1}, &result)
	if err != nil {
		errors.Wrapf(err, "unable to stop the mongos balancer")
	}

	log.Printf(result)
	return nil
}

type resultBalancerStatus struct {
	mode              string `bson:"mode"`
	inBalancerRound   bool   `bson:"inBalancerRound"`
	numBalancerRounds int    `bson:"numBalancerRounds"`
}

func (r resultBalancerStatus) String() string {
	return fmt.Sprintf("mode: %v\n, inBalancerRound: %v\n, numBalancerRounds: %v\n", r.mode, r.inBalancerRound, r.numBalancerRounds)
}

func BalancerStatus(sess *mgo.Session) error {
	var result resultBalancerStatus
	err := sess.Run(bson.M{"balancerStatus": 1}, &result)
	if err != nil {
		errors.Wrapf(err, "unable to get the status of mongos balancer")
	}

	log.Printf(result.String())
	return nil
}
