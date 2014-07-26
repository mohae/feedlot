// builders.go contains all of the builder related functionality for
// rawTemplates. Any new builders should be added here.
package ranchr

import ()

// r.createBuilders takes a raw builder and create the appropriate Packer
// Builders along with a slice of variables for that section builder type.
// Some Settings are in-lined instead of adding them to the variable section.
func (r *rawTemplate) createBuilders() (bldrs []interface{}, vars map[string]interface{}, err error) {
	if r.BuilderTypes == nil || len(r.BuilderTypes) <= 0 {
		err = fmt.Errorf("no builder types were configured, unable to create builders")
		jww.ERROR.Println(err.Error())
		return nil, nil, err
	}

	var vrbls, tmpVar []string
	var tmpS map[string]interface{}
	var k, val, v string
	var i, ndx int
	bldrs = make([]interface{}, len(r.BuilderTypes))

	// Generate the builders for each builder type.
	for _, bType := range r.BuilderTypes {
		jww.TRACE.Println(bType)

		// TODO calculate the length of the two longest Settings and VMSettings sections and make it
		// that length. That will prevent a panic should there be more than 50 options. Besides its
		// stupid, on so many levels, to hard code this...which makes me...d'oh!
		tmpVar = make([]string, 50)
		tmpS = make(map[string]interface{})

		switch bType {
		case BuilderVMWareISO:
//			tmpS, tmpVar, err = r.createBuilderVMWareISO()
		case BuilderVMWareOVF:
//			tmpS, tmpVar, err = r.createBuilderVMWareOVF

		case BuilderVirtualBoxISO:
			tmpS, tmpVar, err = r.createBuilderVirtualBoxISO()

		case BuilderVirtualBoxOVF:
//			tmpS, tmpVar, err = r.createVirtualBoxOVF
			// Generate the common Settings and their vars
			if tmpS, tmpVar, err = r.commonVMSettings(bType, r.Builders[BuilderCommon].Settings, r.Builders[bType].Settings); err != nil {
				jww.ERROR.Println(err.Error())
				return nil, nil, err
			}

			// Generate Packer Variables
			// Generate builder specific section
			tmpVB := make([][]string, len(r.Builders[bType].VMSettings))
			ndx = 0

			for i, v = range r.Builders[bType].VMSettings {
				k, val = parseVar(v)
				val = r.replaceVariables(val)
				tmpVB[i] = make([]string, 4)
				tmpVB[i][0] = "modifyvm"
				tmpVB[i][1] = "{{.Name}}"
				tmpVB[i][2] = "--" + k
				tmpVB[i][3] = val
			}
			tmpS["vboxmanage"] = tmpVB

		default:
			err = errors.New("the requested builder, '" + bType + "', is not supported")
			jww.ERROR.Println(err.Error())
			return nil, nil, err
		}

		tmps["type"] = bType
		bldrs[ndx] = tmpS
		ndx++
		vrbls = append(vrbls, tmpVar...)
	}

	return bldrs, vars, nil
}

/*
// r.createBuilderVMWareISO generates the settings for a vmware-iso builder.
func (r *rawTemplate) createBuilderVMWareISO() (settings map[string]interface{}, vars []string, err error) {
	// Generate the common Settings and their vars
	if tmpS, tmpVar, err = r.commonVMSettings(bType, r.Builders[BuilderCommon].Settings, r.Builders[bType].Settings); err != nil {
		jww.ERROR.Println(err.Error())
		return nil, nil, err
	}

	tmpS["type"] = bType

	// Generate builder specific section
	tmpvm := make(map[string]string, len(r.Builders[bType].VMSettings))

	for i, v = range r.Builders[bType].VMSettings {
		k, val = parseVar(v)
		val = r.replaceVariables(val)
		tmpvm[k] = val
		tmpS["vmx_data"] = tmpvm
	}
}
*/

// r.createBuilderVirtualboxISO generates the settings for a vmware-iso builder.
func (r *rawTemplate) createBuilderVirtualboxISO() (settings map[string]interface{}, vars []string, err error) {
	// Generate the common Settings and their vars
	if tmpS, tmpVar, err = r.commonVMSettings(bType, r.Builders[BuilderCommon].Settings, r.Builders[bType].Settings); err != nil {
		jww.ERROR.Println(err.Error())
		return nil, nil, err
	}

	tmpS["type"] = bType

	// Generate builder specific section
	tmpvm := make(map[string]string, len(r.Builders[bType].VMSettings))

	for i, v = range r.Builders[bType].VMSettings {
		k, val = parseVar(v)
		val = r.replaceVariables(val)
		tmpvm[k] = val
		tmpS["vmx_data"] = tmpvm
	}
}

