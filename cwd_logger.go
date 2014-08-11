package main

import (
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"os"
	"strconv"
	"time"
)

type findType string

const (
	recently   findType = "-last_access"
	frequently findType = "-count"
)

type log struct {
	Id    bson.ObjectId `bson:"_id"`
	Path  string        `bson:"path"`
	Last  time.Time     `bson:"last_access"`
	Count float32       `bson:"count"`
}

type target struct {
	index int
	regex string
}

func logCurrent(c *mgo.Collection) {
	path := os.Getenv("PWD")
	if os.Getenv("HOME") != path {
		c.Upsert(bson.M{"path": path}, bson.M{"$set": bson.M{"path": path, "last_access": time.Now()}, "$inc": bson.M{"count": 1}})
	}
}

func recentlyFrequently(c *mgo.Collection, sort findType) {
	target := getTarget()
	if target.index >= 0 && "" == target.regex {
		printTarget(c, sort, target)
	} else if target.index < 0 && "" != target.regex {
		target.index = 0
		printTarget(c, sort, target)
	} else {
		listRecentyFrequently(c, sort)
	}
}

func printTarget(c *mgo.Collection, sort findType, target target) {
	item := &log{}
	if "" == target.regex {
		c.Find(nil).Sort(string(sort)).Skip(target.index).One(item)
	} else {
		c.Find(bson.M{"path": bson.M{"$regex": target.regex}}).Sort(string(sort)).Skip(target.index).One(item)
	}
	fmt.Println(item.Path)
}

func listRecentyFrequently(c *mgo.Collection, sort findType) {
	iter := c.Find(nil).Sort(string(sort)).Limit(20).Iter()
	item := &log{}
	i := 0
	fmt.Printf("usage: %s target_index|path regex\n", os.Args[0])
	for iter.Next(&item) {
		fmt.Printf("%2d- %s\n", i, item.Path)
		i++
	}
	if err := iter.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Error closing iter: %v\n", err)
	}
}

func getTarget() target {
	target := target{index: -1, regex: ""}
	if len(os.Args) > 1 {
		var err error
		target.index, err = strconv.Atoi(os.Args[1])
		if nil != err {
			target.index = -1
			target.regex = os.Args[1]
		}
	}
	return target
}

func dampenFrequency(c *mgo.Collection) {
	iter := c.Find(nil).Iter()
	item := &log{}
	for iter.Next(&item) {
		item.Count /= 2.0
		c.Update(bson.M{"_id": item.Id}, item)
	}
	if err := iter.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Error closing iter: %v\n", err)
	}
}

func removeDead(c *mgo.Collection) {
	count, err := c.Count()
	if nil != err {
		fmt.Fprintf(os.Stderr, "Can't get collection count: %v\n", err)
		return
	}
	limit := float32(count) * 0.1
	if limit < 20 {
		return
	}
	iter := c.Find(bson.M{"count": bson.M{"$lte": 0.25}}).Sort("last_access").Limit(int(limit)).Iter()
	item := &log{}
	for iter.Next(&item) {
		fmt.Printf("Forgetting about %s\n", item.Path)
		if err := c.Remove(bson.M{"_id": item.Id}); nil != err {
			fmt.Fprintf(os.Stderr, "Error removing item (%v) iter: %v\n", item.Id, err)
		}
	}
	if err := iter.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Error closing iter: %v\n", err)
	}
}

func main() {
	uri := os.Getenv("CWD_LOGGER_URI")
	if "" == uri {
		// mongodb://user:pass@server.mongohq.com/db_name
		uri = "mongodb://localhost/" + os.Getenv("USER")
	}
	sess, err := mgo.DialWithTimeout(uri, 500*time.Millisecond)
	if nil != err {
		fmt.Fprintf(os.Stderr, "Can't connect to mongo: %v\n", err)
		os.Exit(1)
	}
	defer sess.Close()

	collection := sess.DB("").C("logged_dirs")

	switch os.Args[0] {
	case "cwd_logger", "log_cwd", "go_cwd_logger":
		logCurrent(collection)
	case "cwd_recently":
		recentlyFrequently(collection, recently)
	case "cwd_frequency":
		recentlyFrequently(collection, frequently)
	case "cwd_dampen_frequency":
		dampenFrequency(collection)
		removeDead(collection)
	default:
		logCurrent(collection)
	}
}
