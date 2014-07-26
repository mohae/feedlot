// create_provisioners.go creates the provisioners for a Packer build. Add 
// supported provisioners here.
package ranchr

import ()
/*
func (r *rawTemplate) createProvisioners() (p []interface{}, vars map[string]interface{}, err error) {
	if r.ProvisionerType == nil || len(r.ProvisionerType) <= 0 {
		err = fmt.Errorf("no provisioner types were configured, unable to create provisioners")
		jww.ERROR.Println(err.Error())
		return nil, nil, err
	}

	var vrbls, tmpVar []string
	var tmpS map[string]interface{}
	var ndx int
	p = make([]interface{}, len(r.ProvisionerType))

	// Generate the builders for each builder type.
	for _, pType := range r.ProvisionerType {
		jww.TRACE.Println(pType)
		// TODO calculate the length of the two longest Settings sections
		// and make it that length. That will prevent a panic should 
		// there be more than 50 options. Besides its stupid, on so many 
		// levels, to hard code this...which makes me...d'oh!
		tmpVar = make([]string, 50)
		tmpS = make(map[string]interface{})

		switch pType {
		case ProvisionerAnsible:
			// Create the settings
//			tmpS = p.settingsToMap(pType, r)

		case ProvisionerSalt:
			// Create the settings
//			tmpS = p.settingsToMap(pType, r)

		case ProvisionerShellScripts:
			// Create the settings
//			tmpS = p.settingsToMap(pType, r)

		default:
			err = errors.New("the requested provisioner, '" + pType + "', is not supported")
			jww.ERROR.Println(err.Error())
			return nil, nil, err
		}

		p[ndx] = tmpS
		ndx++
		vrbls = append(vrbls, tmpVar...)
	}

	return p, vars, nil
}
*/
