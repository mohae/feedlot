// create_post_processors.go creates the post-processors for a Packer Build. 
// Add supported post-processors here.
package ranchr

import (
	"errors"
	"fmt"
	"strings"

	json "github.com/mohae/customjson"
	jww "github.com/spf13/jwalterweatherman"
)

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
func (r *rawTemplate) updatePostProcessors(new map[string]*postProcessor) {
	// If there is nothing new, old equals merged.
	if len(new) <= 0 || new == nil {
		jww.TRACE.Println("rawTemplate.updatePostProcessors: new was nil Returning w/o doing anything")
		return
	}

	// Convert to an interface.
	var ifaceOld map[string]interface{} = make(map[string]interface{}, len(r.PostProcessors))
	for i, o := range r.PostProcessors {
		ifaceOld[i] = o
	}

	// Do a deep copy of the new postProcessor info.
	var ifaceNew map[string]interface{} = make(map[string]interface{}, len(new))
	ifaceNew = deepCopyMapStringPPostProcessor(new)

	// Get the all keys from both maps
	var keys[]string
	keys = keysFromMaps(ifaceOld, ifaceNew)

//	pM := map[string]postProcessor{}
	p := &postProcessor{}

	for _, v := range keys {
		p = r.PostProcessors[v].DeepCopy()
		
		if p == nil {
			p = &postProcessor{templateSection{Settings: []string{}, Arrays: map[string]interface{}{}}}
		}

		// If the element for this key doesn't exist, skip it.
		if _, ok := new[v]; !ok {
			continue
		}

		p.mergeSettings(new[v].Settings)
		r.PostProcessors[v] = p
	}

	return
}

/*
// Merge the settings section of a post-processor. New values supercede
// existing ones.
func (r *rawTemplate) updateProcessorSettings(new []string) {
	r.updateProcessors(= mergeSettingsSlices(p.Settings, new)
}
*/

// Go through all of the Settings and convert them to a map. Each setting
// is parsed into its constituent parts. The value then goes through
// variable replacement to ensure that the settings are properly resolved.
func (p *postProcessor) settingsToMap(Type string, r *rawTemplate) map[string]interface{} {
	var k string
	var v interface{}
	m := make(map[string]interface{}, len(p.Settings))
	m["type"] = Type

	for _, s := range p.Settings {
		k, v = parseVar(s)

		switch k {
		case "keep_input_artifact":
			if strings.ToLower(fmt.Sprint(v)) == "true" {
				v = true
			} else {
				v = false
			}
		default:
			v = r.replaceVariables(fmt.Sprint(v))
		}

		m[k] = v
	}

	jww.TRACE.Printf("post-processors Map: %v\n", json.MarshalIndentToString(m, "", indent))
	return m
}
// r.createPostProcessors creates the PostProcessors for a build.
func (r *rawTemplate) createPostProcessors() (p []interface{}, vars map[string]interface{}, err error) {
	if r.PostProcessorTypes == nil || len(r.PostProcessorTypes) <= 0 {
		err = fmt.Errorf("no post-processors types were configured, unable to create post-processors")
		jww.ERROR.Println(err.Error())
		return nil, nil, err
	}

	var vrbls, tmpVar []string
	var tmpS map[string]interface{}
	var ndx int
	p = make([]interface{}, len(r.PostProcessorTypes))

	// Generate the builders for each builder type.
	for _, pType := range r.PostProcessorTypes {
		jww.TRACE.Println(pType)
		// TODO calculate the length of the two longest Settings sections
		// and make it that length. That will prevent a panic should 
		// there be more than 50 options. Besides its stupid, on so many 
		// levels, to hard code this...which makes me...d'oh!
		tmpVar = make([]string, 50)
		tmpS = make(map[string]interface{})

		switch pType {
		case PostProcessorVagrant:
			// Create the settings
//			tmpS,  ok = p.(settingsToMap(k, r)

		case PostProcessorVagrantCloud:
			// Create the settings
//			tmpS = p.settingsToMap(k, r)

		default:
			err = errors.New("the requested post-processor, '" + pType + "', is not supported")
			jww.ERROR.Println(err.Error())
			return nil, nil, err
		}

		p[ndx] = tmpS
		ndx++
		vrbls = append(vrbls, tmpVar...)
	}

	return p, vars, nil
}

// deepCopyMapStringPPostProcessor makes a deep copy of each builder passed and 
// returns the copie map[string]*builder as a map[string]interface{}
// notes: This currently only supports string slices.
func deepCopyMapStringPPostProcessor(p map[string]*postProcessor) map[string]interface{} {
	c := map[string]interface{}{}
	for k, v := range p {
		tmpP := &postProcessor{}
		tmpP = v.DeepCopy()
		c[k] = tmpP
	}
	return c
}

