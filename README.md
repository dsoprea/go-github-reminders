[![GoDoc](https://godoc.org/github.com/dsoprea/go-github-reminders?status.svg)](https://godoc.org/github.com/dsoprea/go-github-reminders)


# Overview

This tool produces a daily digest of open Github issues with recent activity that you are subscribed to but haven't recently responded to.


# Search factors

Issue selection involves the following factors:

- You are subscribed to the issue. 
  - Owners, assignees, commentators, and watchers are automatically subscribed.
- The issue is open.
- The issue has been updated within a certain period of time. The default is six months.
- You have not responded within a certain period of time. The default is three days.

To temporarily stop receiving notifications for an issue, respond to it. To indefinitely stop receiving notifications for an issue, unsubscribe from it (in Github).


# Installing

```
$ go get github.com/dsoprea/go-github-reminders/command/check_github
```


# Running

To print the list of reminders in a table in the console using defaults, run:

```
$ "${GOPATH}/bin/check_github" issue-reminders --username <USERNAME> --password <PASSWORD>
```

See command-line help for additional configuration/options.


# Example

```
$ "${GOPATH}/bin/check_github" issue-reminders --username USERNAME --password PASSWORD
2019-01-08 04:03:20 +0000 UTC https://api.github.com/repos/dsoprea/go-exif/issues/7 dsoprea/go-exif mysterytree Wrong latlong
2019-01-04 00:11:29 +0000 UTC https://api.github.com/repos/dsoprea/go-exif/issues/1 dsoprea/go-exif evanoberholster Question about MakerNotes
2018-11-08 19:53:10 +0000 UTC https://api.github.com/repos/dsoprea/PySvn/issues/95 dsoprea/PySvn ghost Return output from svn command
2018-10-02 09:56:39 +0000 UTC https://api.github.com/repos/dsoprea/PyInotify/issues/48 dsoprea/PyInotify xlotlu use os.walk for recursing
2018-09-21 08:57:43 +0000 UTC https://api.github.com/repos/dsoprea/PySvn/issues/102 dsoprea/PySvn h0h0h0 svn import functionality?
2018-09-20 00:40:32 +0000 UTC https://api.github.com/repos/dsoprea/PyInotify/issues/30 dsoprea/PyInotify Larivact Add ignored_dirs param to InotifyTree(s)
2018-08-22 15:38:59 +0000 UTC https://api.github.com/repos/dsoprea/PySvn/issues/123 dsoprea/PySvn tmzhuang Pledgie link on PyPi is down
2018-08-09 15:46:14 +0000 UTC https://api.github.com/repos/dsoprea/GDriveFS/issues/165 dsoprea/GDriveFS sketch242 Keepass database
2018-07-30 01:49:09 +0000 UTC https://api.github.com/repos/dsoprea/PySvn/issues/119 dsoprea/PySvn matt4d617474 Support LocalClient.{delete,move} operations (expand testing too)
2018-07-26 13:58:41 +0000 UTC https://api.github.com/repos/dsoprea/PyInotify/issues/31 dsoprea/PyInotify innlym Fixed bug: mkdir -p foo/bar,InotifyTree(s) not add watch bar
2018-07-25 15:55:26 +0000 UTC https://api.github.com/repos/dsoprea/PyInotify/issues/50 dsoprea/PyInotify Beefster09 Context Manager?
```

There are a couple of recent issues that need a response, as well as a few older ones needing follow-up.


# Scheduling Execution

Scheduling is not managed by the tool. Just use [Cron](https://en.wikipedia.org/wiki/Cron).
