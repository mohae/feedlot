// create_post_processors.go creates the post-processors for a Packer Build. 
// Add supported post-processors here.
package ranchr

import (
	"errors"
	"fmt"

	jww "github.com/spf13/jwalterweatherman"
)

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
