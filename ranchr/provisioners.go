// create_provisioners.go creates the provisioners for a Packer build. Add 
// supported provisioners here.
package ranchr

import ()

// Merges the new config with the old. The updates occur as follows:
//
//	* The existing configuration is used when no `new` postProcessors are
//	  specified.
//	* When 1 or more `new` postProcessors are specified, they will replace
//        all existing postProcessors. In this situation, if a postProcessor
//	  exists in the `old` map but it does not exist in the `new` map, that
//        postProcessor will be orphaned.
// If there isn't a new config, return the existing as there are no
// overrides
func (r *rawTemplate) updateProvisioners(new map[string]*provisioner) {
	// If there is nothing new, old equals merged.
	if len(new) <= 0 || new == nil {
		return
	}

	// Convert to an interface.
	var ifaceOld map[string]interface{} = make(map[string]interface{}, len(r.Provisioners))
	for i, o := range r.Provisioners {
		ifaceOld[i] = o
	}

	// Convert to an interface.
	var ifaceNew map[string]interface{} = make(map[string]interface{}, len(new))
	for i, n := range new {
		ifaceNew[i] = n
	}

	// Get the all keys from both maps
	var keys[]string
	keys = keysFromMaps(ifaceOld, ifaceNew)

	pM := map[string]provisioner{}
	p := &provisioner{}

	for _, v := range keys {
		p = r.Provisioners[v]
		
		if p == nil {
			p = &provisioner{templateSection{Settings: []string{}, Arrays: map[string]interface{}{}}}
		}

		// If the element for this key doesn't exist, skip it.
		if _, ok := new[v]; !ok {
			continue
		}

		p.mergeSettings(new[v].Settings)
		pM[v] = *p
	}

	return
}

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
