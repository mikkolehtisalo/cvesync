Cvesync
=======

Introduction
------------

Accidentally disregarding known information-security vulnerabilities and exposures may lead to dire consequences. Tracking CVEs reliably requires great amount of work. Cvesync assists in previous by synchronizing new CVEs to an issue management system. After that the workflow included within issue management system can assist in the analysis, mitigation, and patching.

By default cvesync reads the modified feed provided by [nvd](https://nvd.nist.gov), and updates Jira. The outcome looks something like [this](https://raw.githubusercontent.com/mikkolehtisalo/cvesync/master/jira.png).

Installation
------------

The following prerequisities should be met:

* Golang 1.3+
* sqlite3
* [go-sqlite3|github.com/mattn/go-sqlite3]
* [blackjack/syslog|ithub.com/blackjack/syslog]
* Jira

Cvesync can be built and installed with make:

```sh
go get github.com/mikkolehtisalo/cvesync
...
make
sudo make install
```

Configuration
-------------

The common options can be found from /opt/cvesync/etc/settings.json:

```json
{
    "CAKeyFile": "/opt/cvesync/etc/ca.crt",
    "FeedURL": "https://nvd.nist.gov/feeds/xml/cve/nvdcve-2.0-Modified.xml.gz",
    "CWEfile": "/opt/cvesync/etc/cwec_v2.8.xml",
    "DBFile": "/opt/cvesync/var/cvesync.sqlite"
}
```

The CAKeyFile points to CA Certificate chain that is used for validating the NVD's server. Before you run cvesync you should verify that it and the used URL are valid.

Jira specific options can be found from /opt/cvesync/etc/jira.json:

```json
{
    "BaseURL": "http://dev.localdomain:8080",
    "Username": "admin",
    "Password": "password",
    "Project": "10000",
    "Issuetype": "10000",
    "TemplateFile": "/opt/cvesync/etc/jira.templ", 
    "HighPriority": "2",
    "MediumPriority": "3",
    "LowPriority": "4"
}
```

It is recommended that you create separate user, project, priorities, and issue type in Jira. Also it is recommendable to evaluate different workflows for the vulnerability issue type. Also, make sure that the description field renderer is Wiki Style Renderer instead of Default Text Renderer.

SELinux
-------

A simple SELinux policy is included. To install it, use make:

```sh
sudo make selinux
```

Running
-------

NVD's CVE feeds update at maximum once per two hours. Cvesync should most likely be run daily via cron, for example:

```sh
0 5 * * * /opt/cvesync/bin/cvesync
```

Notes
-----

* NVD recommends that the CVEs are classified with scale Low-Medium-High. Vulnerabilities with a base score in the range 7.0-10.0 are High, those in the range 4.0-6.9 as Medium, and 0-3.9 as Low.
* CWE xml can be downloaded from http://cwe.mitre.org/data/index.html#downloads . It doesn't update very often.
* There is an interface (*Tracker*) for implementing other issue management systems
* Logging is done to syslog facility DAEMON. If it is not meaningful to recover, the application panics.
* Connection to Jira is by default HTTP
