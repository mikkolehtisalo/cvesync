Cvesync
=======

Introduction
------------

Accidentally disregarding known information-security vulnerabilities and exposures may lead to dire consequences. Tracking CVEs reliably requires great amount of work. Cvesync assists in previous by synchronizing new CVEs to an issue management system. After that the workflow included within issue management system can assist in the analysis, mitigation, and patching.

By default cvesync reads the modified feed provided my [nvd|https://nvd.nist.gov], and updates Jira. 

Installation
------------

The following prerequisities should be met:

* Golang 1.3+
* sqlite3
* [go-sqlite3|github.com/mattn/go-sqlite3]
* [blackjack/syslog|ithub.com/blackjack/syslog]

Cvesync can be built and installed with make:

```sh
make
sudo make install
```

Configuration
-------------

The common options can be found from /opt/cvesync/etc/settings.json:

```json
{
    "CAKeyFile": "/opt/cvesync/etc/ca.crt",  // CA certificate chain for the following
    "FeedURL": "https://nvd.nist.gov/feeds/xml/cve/nvdcve-2.0-Modified.xml.gz", // Where the CVE feed is fetched from
    "CWEfile": "/opt/cvesync/etc/cwec_v2.8.xml", // The file with CWE information, must be updated manually
    "DBFile": "/opt/cvesync/var/cvesync.sqlite" // Database for tracking synchronization status
}

Before you run cvesync you should at minimum verify the CA certificate chain, and the feed url.

Jira specific options can be found from /opt/cvesync/etc/jira.json:

```json
{
    "BaseURL": "http://dev.localdomain:8080",
    "Username": "admin",
    "Password": "password",
    "Project": "10000",
    "Issuetype": "10000",
    "TemplateFile": "/opt/cvesync/etc/jira.templ", // Golang text/template for description
    "HighPriority": "2", // For mapping to custom priorities
    "MediumPriority": "3", // For mapping to custom priorities
    "LowPriority": "4" // For mapping to custom priorities
}
```

It is recommended that you create separate user, project, and issue type in Jira. Also it is recommendable to evaluate different workflows for the vulnerability issue type. Also, make sure that the description field renderer is Wiki Style Renderer instead of Default Text Renderer.

SELinux
-------

A simple SELinux policy is included. To install it, use make:

```sh
sudo make selinux
```

Notes
-----

* NVD recommends that the CVEs are classified with scale Low-Medium-High. Vulnerabilities with a base score in the range 7.0-10.0 are High, those in the range 4.0-6.9 as Medium, and 0-3.9 as Low.
* CWE xml can be downloaded from http://cwe.mitre.org/data/index.html#downloads
* There is an interface (*Tracker*) for implementing other issue management systems
* Logging is done to syslog facility DAEMON. If it is not meaningful to recover, the application panics.
* Connection to Jira is by default HTTP
