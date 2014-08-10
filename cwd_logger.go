package main

import (
    "fmt"
    "os"
    "time"
    "labix.org/v2/mgo"
    "labix.org/v2/mgo/bson"
    "strconv"
)

type FindType string
const (
    Recently FindType = "-last_access"
    Frequently FindType = "-count"
)

type Log struct {
    Id    bson.ObjectId `bson:"_id"`
    Path  string        `bson:"path"`
    Last  time.Time     `bson:"last_access"`
    Count float32       `bson:"count"`
}

func LogCurrent(c *mgo.Collection){
    path := os.Getenv("PWD")
    if os.Getenv("HOME") != path {
        c.Upsert(bson.M{"path": path}, bson.M{"$set": bson.M{"path": path, "last_access": time.Now()}, "$inc":bson.M{"count":1}})
    }
}

func RecentlyFrequently(c *mgo.Collection, sort FindType){
    target := getTarget()
    if target < 0 {
        listRecentyFrequently(c,sort)
    } else {
        printTarget(c,sort,target)
    }
}

func printTarget(c *mgo.Collection, sort FindType, target int) {
    item := &Log{}
    c.Find(nil).Sort(string(sort)).Skip(target).One(item)
    fmt.Println(item.Path)
}

func listRecentyFrequently(c *mgo.Collection, sort FindType) {
    iter := c.Find(nil).Sort(string(sort)).Limit(20).Iter()
    item := &Log{}
    i := 0
    fmt.Printf("usage: %s target_index\n",os.Args[0])
    for iter.Next(&item) {
        fmt.Printf("%2d- %s\n",i,item.Path)
        i+=1
    }
    if err := iter.Close(); err != nil {
        fmt.Fprintf(os.Stderr,"Error closing iter: %v\n",err)
    }
}

func getTarget() int {
    if len(os.Args) > 1 {
        target, err := strconv.Atoi(os.Args[1])
        if nil != err {
            fmt.Fprintf(os.Stderr,"Error parsing target: %v\n",err)
            return -1
        }
        return target
    }
    return -1
}

func DampenFrequency(c *mgo.Collection) {
    iter := c.Find(nil).Iter()
    item := &Log{}
    for iter.Next(&item) {
        item.Count /= 2.0
        c.Update(bson.M{"_id": item.Id},item)
    }
    if err := iter.Close(); err != nil {
        fmt.Fprintf(os.Stderr,"Error closing iter: %v\n",err)
    }
}

func RemoveDead(c *mgo.Collection) {
    count, err := c.Count()
    if nil != err {
        fmt.Fprintf(os.Stderr,"Can't get collection count: %v\n",err)
        return
    }
    limit := float32(count) * 0.1
    if limit < 20 {
        return
    }
    iter := c.Find(bson.M{"count": bson.M{"$lte": 0.25}}).Sort("last_access").Limit(int(limit)).Iter()
    item := &Log{}
    for iter.Next(&item) {
        fmt.Printf("Forgetting about %s\n",item.Path)
        if err := c.Remove(bson.M{"_id": item.Id}); nil != err {
            fmt.Fprintf(os.Stderr,"Error removing item (%v) iter: %v\n",item.Id,err)            
        }
    }
    if err := iter.Close(); err != nil {
        fmt.Fprintf(os.Stderr,"Error closing iter: %v\n",err)
    }
    
}

func main () {
    uri := os.Getenv("CWD_LOGGER_URI")
    if "" == uri {
        // mongodb://user:pass@server.mongohq.com/db_name
        uri = "mongodb://localhost/" + os.Getenv("USER")
    }
    sess, err := mgo.DialWithTimeout(uri,500*time.Millisecond)
    if nil != err {
        fmt.Fprintf(os.Stderr,"Can't connect to mongo: %v\n", err)
        os.Exit(1)
    }
    defer sess.Close()

    collection := sess.DB("").C("logged_dirs")

    switch os.Args[0] {
    case "cwd_logger", "log_cwd", "go_cwd_logger":
        LogCurrent(collection)
    case "cwd_recently":
        RecentlyFrequently(collection, Recently)
    case "cwd_frequency":
        RecentlyFrequently(collection, Frequently)
    case "cwd_dampen_frequency":
        DampenFrequency(collection)
        RemoveDead(collection)
    default:
        LogCurrent(collection)
    }
}
