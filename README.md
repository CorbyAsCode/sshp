# sshp - parallel ssh when you want just text output
This is a binary executable that will execute commands over ssh to multiple hosts at the same time.
It will only run on 10 hosts at any one time.
There are multiple ways to specify what hosts you want to ssh into.
It will only output the result of your command or the error that was encountered.

This has only been compiled and tested for Linux x86_64 using GOOS=Linux GOARCH=amd64.

## Examples
    [corby@homebase ~]# sshp -h
    Usage of sshp:
    -c string
            Command to execute over ssh.
    -config string
            Config file to use. (default "/etc/sshp.conf")
    -d	Print debug messages.
    -f string
            Host file to use, one hostname per line.
    -h	Print help.
    -iniFile string
            Path to ini file.
    -iniSection string
            ini section to retrieve.
    -k string
            Path to ssh key.
    -l string
            Comma-separated list of hosts.
    -s	Use sudo to become root on hosts.
    -u string
            User to login as.

### Specify comma-separated list of hosts on the CLI
    [corby@homebase ~]# sshp -l host01,host02 -c "uname -a"
    host01: Linux host01 2.6.32-754.3.5.el6.x86_64 #1 SMP Thu Aug 9 11:56:22 EDT 2018 x86_64 x86_64 x86_64 GNU/Linux
    host02: Linux host02 2.6.32-754.3.5.el6.x86_64 #1 SMP Thu Aug 9 11:56:22 EDT 2018 x86_64 x86_64 x86_64 GNU/Linux

### Use a list of hosts in a file
    [corby@homebase ~]# cat /tmp/sshp_hosts 
    host01
    host02
    host03
    host04
    host05
    
    [corby@homebase ~]# sshp -f /tmp/sshp_hosts -c "uname -a"
    host01: Linux host01 2.6.32-754.3.5.el6.x86_64 #1 SMP Thu Aug 9 11:56:22 EDT 2018 x86_64 x86_64 x86_64 GNU/Linux
    host02: Linux host02 2.6.32-754.3.5.el6.x86_64 #1 SMP Thu Aug 9 11:56:22 EDT 2018 x86_64 x86_64 x86_64 GNU/Linux
    host03: Linux host03 2.6.32-754.3.5.el6.x86_64 #1 SMP Thu Aug 9 11:56:22 EDT 2018 x86_64 x86_64 x86_64 GNU/Linux
    host04: Linux host05 2.6.32-754.3.5.el6.x86_64 #1 SMP Thu Aug 9 11:56:22 EDT 2018 x86_64 x86_64 x86_64 GNU/Linux
    host06: Linux host06 2.6.32-754.3.5.el6.x86_64 #1 SMP Thu Aug 9 11:56:22 EDT 2018 x86_64 x86_64 x86_64 GNU/Linux

### Specify a particular section of an ini file...in this example, the Ansible hosts file
    [corby@homebase ~]# sshp -iniFile /etc/ansible/hosts -iniSection rhel6 -c "uname -r"
    host01: 2.6.32-754.3.5.el6.x86_64
    host02: 2.6.32-754.3.5.el6.x86_64
    host03: 2.6.32-754.3.5.el6.x86_64
    host04: 2.6.32-754.3.5.el6.x86_64
    host05: 2.6.32-754.3.5.el6.x86_64
    host06: 2.6.32-754.3.5.el6.x86_64
    host07: 2.6.32-754.3.5.el6.x86_64
    host08: 2.6.32-754.3.5.el6.x86_64
    host09: 2.6.32-754.3.5.el6.x86_64
    host10: 2.6.32-754.3.5.el6.x86_64
    host11: 2.6.32-754.3.5.el6.x86_64
    host12: 2.6.32-754.3.5.el6.x86_64

### You can use -s for sudo...
    [corby@homebase ~]$ sshp -iniFile /etc/ansible/hosts -iniSection stress -s -c "service ntpd status"
    host01: ntpd (pid  151) is running...
    host02: ntpd (pid  45326) is running...
    host03: ntpd (pid  22661) is running...

