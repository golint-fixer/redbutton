set -e
ansible-playbook playbooks/build.yml

inventory/ec2.py --refresh-cache > /dev/null
ansible-playbook playbooks/tag_for_decomission.yml

ansible-playbook playbooks/provision.yml

inventory/ec2.py --refresh-cache > /dev/null
ansible-playbook playbooks/install-application.yml
ansible-playbook playbooks/decomission.yml

