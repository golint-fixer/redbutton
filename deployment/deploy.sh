ansible-playbook playbooks/build.yml
ansible-playbook playbooks/decomission.yml

inventory/ec2.py --refresh-cache > /dev/null
ansible-playbook playbooks/provision.yml

inventory/ec2.py --refresh-cache > /dev/null
ansible-playbook playbooks/deploy.yml