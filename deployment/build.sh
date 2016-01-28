source ../venv/bin/activate
ansible-playbook playbooks/build.yml
ansible-playbook playbooks/deploy.yml
