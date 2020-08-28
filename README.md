# k3ama - Air-gap Migration Assistant (working name)
```bash
#  _    _____   _    __  __    _
# | | _|___ /  / \  |  \/  |  / \
# | |/ / |_ \ / _ \ | |\/| | / _ \
# |   < ___) / ___ \| |  | |/ ___ \
# |_|\_\____/_/   \_\_|  |_/_/   \_\ k3s- Air-gap Migration Assistant
#
#                ,        ,  _______________________________
#    ,-----------|'------'|  |                             |
#   /.           '-'    |-'  |_____________________________|
#  |/|             |    |
#    |   .________.'----'    _______________________________
#    |  ||        |  ||      |                             |
#    \__|'        \__|'      |_____________________________|
#
# |‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾|
# |________________________________________________________|
#                                                          |
# |‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾|
# |________________________________________________________|

```
WARNING- Work In Progress

## Prerequisites
* CentOS 7
* User with root/sudo privileges
* 

## Installing on an airgap network
1) (Skip if you aren't using SELINUX) Install the `selinux` dependencies. `yum localinstall -y ./artifacts/yum/*`.
2) For some reason, centos doesn't add `/usr/local/bin` to the path. Add it with `echo 'export PATH=${PATH}:/usr/local/bin' >> ~/.bashrc`
3) 



## Charts to include
* Rancher
* Registry
* Minio
* Longhorn
* git-http-backend
* argo

## TODO
* Write the thing
* Include Vagrantfile for testing

### Other possible names
* k3vac
* k3ziplock
* k3wh - k3 wormhole
* k3cia - Comms insensensitive Assistant
* k3diode