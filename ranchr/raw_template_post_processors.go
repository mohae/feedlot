// create_post_processors.go creates the post-processors for a Packer Build.
// Add supported post-processors here.
package ranchr

import (
	"errors"
	"fmt"
	"strings"

	json "github.com/mohae/customjson"
	"github.com/mohae/deepcopy"
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

	// Convert the existing postProcessors to interface.
	var ifaceOld map[string]interface{} = make(map[string]interface{}, len(r.PostProcessors))
	ifaceOld = DeepCopyMapStringPPostProcessor(r.PostProcessors)
	//	for i, o := range r.PostProcessors {
	//		ifaceOld[i] = o
	//	}

	// Convert the new postProcessors to interfaces
	var ifaceNew map[string]interface{} = make(map[string]interface{}, len(new))
	ifaceNew = DeepCopyMapStringPPostProcessor(new)

	// Get the all keys from both maps
	var keys []string
	keys = mergedKeysFromMaps(ifaceOld, ifaceNew)
	p := &postProcessor{}

	if r.PostProcessors == nil {
		r.PostProcessors = map[string]*postProcessor{}
	}

	// Copy: if the key exists in the new postProcessors only.
	// Ignore: if the key does not exist in the new postProcessors.
	// Merge: if the key exists in both the new and old postProcessors.
	for _, v := range keys {
		// If it doesn't exist in the old builder, add it.
		_, ok := r.PostProcessors[v]
		if !ok {
			r.PostProcessors[v] = new[v].DeepCopy()
			continue
		}
		// If the element for this key doesn't exist, skip it.
		_, ok = new[v]
		if !ok {
			continue
		}
		p = r.PostProcessors[v].DeepCopy()
		if p == nil {
			p = &postProcessor{templateSection{Settings: []string{}, Arrays: map[string]interface{}{}}}
		}
		p.mergeSettings(new[v].Settings)
		p.mergeArrays(new[v].Arrays)
		r.PostProcessors[v] = p
	}
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
		// If its not set to 'true' then false. This is a case insensitive
		// comparison.
		// TODO why am I using fmt.Sprint(v) here?
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
		err = errors.New("unable to create post-processors: none specified")
		jww.ERROR.Println(err)
		return nil, nil, err
	}
	var vrbls, tmpVar []string
	var tmpS map[string]interface{}
	var ndx int
	p = make([]interface{}, len(r.PostProcessorTypes))

	// Generate the postProcessor for each postProcessor type.
	for _, pType := range r.PostProcessorTypes {
		jww.TRACE.Println(pType)
		// TODO calculate the length of the two longest Settings sections
		// and make it that length. That will prevent a panic unless
		// there are more than 50 options. Besides its stupid, on so many
		// levels, to hard code this...which makes me...d'oh!
		tmpVar = make([]string, 50)
		tmpS = make(map[string]interface{})
		typ := PostProcessorFromString(pType)
		switch typ {
		case Vagrant:
			tmpS, tmpVar, err = r.createVagrant()
		case VagrantCloud:
			// Create the settings
			tmpS, tmpVar, err = r.createVagrantCloud()
		default:
			err = fmt.Errorf("%s is not supported", pType)
			jww.ERROR.Println(err)
			return nil, nil, err
		}
		p[ndx] = tmpS
		ndx++
		vrbls = append(vrbls, tmpVar...)
	}
	return p, vars, nil
}

// createVagrant() creates a map of settings for Packer's Vagrant post-processor.
// Any values that aren't supported by the Vagrant post-processor are ignored.
func (r *rawTemplate) createVagrant() (settings map[string]interface{}, vars []string, err error) {
	settings = make(map[string]interface{})
	settings["type"] = Vagrant

	jww.TRACE.Printf("rawTemplate.createPostProcessorVagrant-rawtemplate\n")

	// For each value, extract its key value pair and then process. Only
	// process the supported keys. Key validation isn't done here, leaving
	// that for Packer.
	var k, v string
	for _, s := range r.PostProcessors[Vagrant.String()].Settings {
		k, v = parseVar(s)
		switch k {
		case "compression_level", "keep_input_artifact", "output":
			settings[k] = v
		default:
			jww.WARN.Println("An unsupported key was encountered: " + k)
		}
	}
	// Process the Arrays.
	for name, val := range r.PostProcessors[Vagrant.String()].Arrays {
		jww.TRACE.Printf("Arrays:\t%v\t%v\n\n", name, val)
		array := deepcopy.Iface(val)
		if array != nil {
			settings[name] = array
		}
	}
	return settings, vars, nil
}

func (r *rawTemplate) createVagrantCloud() (settings map[string]interface{}, vars []string, err error) {
	settings = make(map[string]interface{})
	settings["type"] = VagrantCloud
	return nil, nil, nil
}

// DeepCopyMapStringPPostProcessor makes a deep copy of each builder passed and
// returns the copie map[string]*builder as a map[string]interface{}
// notes: This currently only supports string slices.
func DeepCopyMapStringPPostProcessor(p map[string]*postProcessor) map[string]interface{} {
	c := map[string]interface{}{}
	for k, v := range p {
		tmpP := &postProcessor{}
		tmpP = v.DeepCopy()
		c[k] = tmpP
	}
	return c
}
