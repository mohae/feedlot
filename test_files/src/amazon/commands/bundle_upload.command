sudo -n ec2-bundle-vol -k {{.KeyPath}} -u {{.AccountId}} -c {{.CertPath}} -r {{.Architecture}} -e {{.PrivatePath}} -d {{.Destination}} -p {{.Prefix}} --batch --no-filter
