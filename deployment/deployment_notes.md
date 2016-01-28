High level instructions for performing deployment to AWS.

## setting up

* create and activate python virtual environment
  ```
  $ virtualenv venv
  $ source venv/bin/activate
  ```
* Install Ansible & Boto
  ```
  $ pip install ansible
  $ pip install boto
  ```
* configure AWS API credentials in Boto - `~/.boto` file worked fine:
  ```
  [Credentials]
  aws_access_key_id = ...
  aws_secret_access_key = ...
  ```
* configure AWS instance SSH with private key access:
  * create key pair `redbutton` in AWS console (or change name in `deploy.yml`)
  * add keypair to ssh:
    ```
    $ chmod u=rw,g-rwx,o-rwx ~/.ssh/redbutton.pem
    $ ssh-add ~/.ssh/redbutton.pem
    ```

## running deployments

* make sure virtual env is activated
* deploy:

  ```
  ansible-playbook playbooks/deploy.yml
  ```

## checking
Ping launched instances:

```
ansible -m ping -u ubuntu ec2
```