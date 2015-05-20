curl -L https://www.opscode.com/chef/install.sh | {{if .Sudo}}sudo{{end}} bash
