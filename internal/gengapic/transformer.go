package gengapic

import (
	"fmt"
	//"net/url"
	"strings"
)

type ItemSdk struct {
	mapPart map[interface{}]interface{}
	name string
	isRepeated bool
	isEnum bool
	inSpec string
	functionName string
}
func (g *generator) generateTransFormerOut(s map[interface{}]interface{}, functionName string, specName string, inSpec string) []ItemSdk {
	p := g.printf

	sdkMapping := getMetaData(s["metadata"]).mapping
	sdkName := specInputToParam(sdkMapping)

	returnList := []ItemSdk{}
	p("func %s (req *%s.%s)  *%s {", functionName, specName, sdkName, inSpec)

	p("if req == nil {")
	p(" return %s{}", inSpec)
	p("}")

	p("wssd%s := &%s{", inSpec, inSpec)

	for kk, v := range s {
		if kk == "metadata" {
			continue
		}
		stringval, ok := v.(string)
		if !ok {
			shouldwork := v.(map[interface{}]interface{})
//			sdkItem := shouldwork["sdkmapping"].(string)
			metaData := getMetaData(shouldwork["metadata"])

			if metaData.mapping == "" {
				continue
			}
			p("%s: getSDK%s(req.%s),", kk, kk, metaData.mapping)
			
			functionName := fmt.Sprintf("getSDK%s", kk)

			if g.funcMap[functionName] == true {
				continue
			}
			g.funcMap[functionName] = true

			returnList = append(returnList, ItemSdk{
					mapPart: shouldwork,
					name: kk.(string),
					isRepeated: metaData.isRepeated,
					functionName: functionName,
			})
		} else {
			if stringval != "" {
				p("%s: req.%s,", kk, stringval)	
			}
		}
	}
	p("}")
	p("	return wssd%s", inSpec) 
	p("}")

	return returnList
}

type MetaData struct {
	isRepeated bool
	mapping string
	inSpec string
}

func getMetaData(md interface{}) MetaData {
	meta := md.(map[interface{}]interface{})
	mdObject := MetaData{
		isRepeated: meta["repeated"].(bool),
		mapping: meta["mapping"].(string),
	}
	inspec, ok := meta["inSpec"].(string)
	if ok {
		mdObject.inSpec = inspec
	}
	return mdObject
}

func (g *generator) generateTransFormerRepeatedOut(s map[interface{}]interface{}, functionName string, specName string, inSpec string) []ItemSdk {
	p := g.printf

	sdkMapping := getMetaData(s["metadata"]).mapping
	sdkName := specInputToParam(sdkMapping)
	sdkNameSingle := strings.TrimSuffix(sdkName, "s")
	inSpecSingle := strings.TrimSuffix(inSpec, "s")
	iter := lowerFirst(sdkNameSingle)
	returnList := []ItemSdk{}


	p("func %s (req []*%s.%s)  []*%s {", functionName, specName, sdkNameSingle, inSpecSingle)
	p("wssd%s := []*%s{}", inSpecSingle, inSpecSingle)

	p("for _, %s := range req {", iter)
	p("  wssd%s = append(wssd%s, &%s{", inSpecSingle, inSpecSingle, inSpecSingle)
	for kk, v := range s {
		if kk == "metadata" {
			continue
		}
		stringval, ok := v.(string)
		if !ok {
			shouldwork := v.(map[interface{}]interface{})
			metaData := getMetaData(shouldwork["metadata"])
			//sdkNameForMeth := specInputToParam(sdkItem)

			if metaData.mapping == "" {
				continue
			}

			if iter == lowerFirst(metaData.mapping) {
				p("%s: getSDK%s(%s),", kk, kk, iter)
			} else {
				p("%s: getSDK%s(%s.%s),", kk, kk, iter, metaData.mapping)
			}
			//isPassThru := false
			//if len(shouldwork) == 2 {
			//	isPassThru = true
			//}

			
			functionName := fmt.Sprintf("getSDK%s", kk)

			if g.funcMap[functionName] == true {
				continue
			}
			g.funcMap[functionName] = true

			returnList = append(returnList, ItemSdk{
					mapPart: shouldwork,
					name: kk.(string),
					isRepeated: metaData.isRepeated,
					functionName: functionName,
			})
			
		} else {
			if stringval != "" {
				p("%s: %s.%s,", kk, iter, stringval)	
			}
		}
	}

	p("})")
	p("}")
	p("	return wssd%s", inSpecSingle) 
	p("}")

	return returnList
}

func (g *generator) generateTransFormer(s map[interface{}]interface{}, functionName string, specName string, inSpec string) []ItemSdk {
	p := g.printf

	sdkMapping := getMetaData(s["metadata"]).mapping
	isRepeatedToplevel := getMetaData(s["metadata"]).isRepeated
	sdkName := specInputToParam(sdkMapping)
	returnList := []ItemSdk{}
	p("func %s (req *%s)  *%s.%s {", functionName, sdkName, specName, inSpec)

	p("if req == nil {")
	p(" return %s.%s{}", specName, inSpec)
	p("}")

	p("wssd%s := &%s.%s{", inSpec, specName, inSpec)

	for kk, v := range s {
		if kk == "metadata" {
			continue
		}

		stringval, ok := v.(string)
		if !ok {
			shouldwork := v.(map[interface{}]interface{})
			metaData := getMetaData(shouldwork["metadata"])
			var param string

			if metaData.mapping == "" {
				continue
			}

			if metaData.isRepeated  && !isRepeatedToplevel {
				param = fmt.Sprintf("[]*%s{req}", sdkName)
			} else {
				param = fmt.Sprintf("req")
			}

			isEnum := false
			if shouldwork["enum"] != nil {
				isEnum = true
			}

			functionName := fmt.Sprintf("getWssd%s", kk)
			if isEnum {
				functionName = fmt.Sprintf("getWssd%s", shouldwork["enum"].(string))
			} 
			
			if kk.(string) == metaData.mapping {
				// special case past through
				p("%s: %s(%s),", kk, functionName, param)
			} else {
				p("%s: %s(req.%s),", kk, functionName, metaData.mapping)
			}

			if g.funcMap[functionName] == true {
				continue
			}
			g.funcMap[functionName] = true

			returnList = append(returnList, ItemSdk{
				mapPart: shouldwork,
				name: kk.(string),
				isRepeated: metaData.isRepeated,
				isEnum: isEnum,
				functionName: functionName,
		})
		} else {
			if stringval != "" {
				p("%s: req.%s,", kk, stringval)	
			}
		}
	}

	p("}")
	p("	return wssd%s", inSpec) 
	p("}")

	return returnList
}

func (g *generator) generateTransFormerEnum(s map[interface{}]interface{}, functionName string, specName string, inSpec string) {
	p := g.printf

	//sdkMapping := s["sdkmapping"]
	enumName := s["enum"].(string)

	//sdkStr := sdkMapping.(string)
	//sdkName := specInputToParam(sdkStr)
	p("func %s (enum string)  %s.%s {", functionName, specName, enumName)

	p("typevalue := %s.%s(0)", specName, enumName)
	p("typevTmp, ok := %s.%s_value[enum]", specName, enumName)
	p("if ok {")
		p("typevalue = %s.%s(typevTmp)", specName, enumName)
	p("}")
	p("	return typevalue")
	p("}")
}

func (g *generator) generateTransFormerRepeated(s map[interface{}]interface{}, functionName string, specName string, inSpec string) []ItemSdk {
	p := g.printf

	sdkMapping := getMetaData(s["metadata"]).mapping
	sdkName := specInputToParam(sdkMapping)
	sdkNameSingle := strings.TrimSuffix(sdkName, "s")
	inSpecSingle := strings.TrimSuffix(inSpec, "s")
	iter := lowerFirst(inSpecSingle)
	returnList := []ItemSdk{}
	p("func %s (req []*%s)  []*%s.%s {", functionName, sdkNameSingle, specName, inSpecSingle)

	p("wssd%s := []*%s.%s{}", inSpecSingle, specName, inSpecSingle)

	p("for _, %s := range req {", iter)
	p("  wssd%s = append(wssd%s, &%s.%s{", inSpecSingle, inSpecSingle, specName, inSpecSingle)
	for kk, v := range s {
		if kk == "metadata" {
			continue
		}
		stringval, ok := v.(string)
		if !ok {
			shouldwork := v.(map[interface{}]interface{})
			metaData := getMetaData(shouldwork["metadata"])

			if metaData.mapping == "" {
				continue
			}
			//sdkItem := shouldwork["sdkmapping"].(string)
			//sdkNameForMeth := specInputToParam(sdkItem)

			//if kk.(string) == sdkName {
				// special case past through
		//		p("%s: getWssd%s(req),", kk, kk)
		//	} else {
			//}
			isEnum := false
			if shouldwork["enum"] != nil {
				isEnum = true
			}

			functionName := fmt.Sprintf("getWssd%s", kk)
			if isEnum {
				functionName = fmt.Sprintf("getWssd%s", shouldwork["enum"].(string))
			}

			p("%s: %s(%s.%s),", kk, functionName, iter, metaData.mapping)

			if g.funcMap[functionName] == true {
				continue
			}
			g.funcMap[functionName] = true
			returnList = append(returnList, ItemSdk{
				mapPart: shouldwork,
				name: kk.(string),
				isRepeated: metaData.isRepeated,
				isEnum: isEnum,
				functionName: functionName, 
			})
		} else {
			p("%s: %s.%s,", kk, iter, stringval)
		}
	}
	p("})")
	p("}")
	p("	return wssd%s", inSpecSingle) 
	p("}")
	return returnList
}

func specInputToParam(specin string) string {
	out := strings.Split(specin, ".")
	return out[len(out)-1]
}
