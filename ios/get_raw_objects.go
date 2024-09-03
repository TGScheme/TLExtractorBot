package ios

import (
	"TLExtractor/asm"
	"TLExtractor/environment"
	"TLExtractor/ios/types"
	"TLExtractor/telegram/scheme"
	"errors"
	"fmt"
	"github.com/Laky-64/gologging"
	"github.com/Laky-64/goswift/demangling"
	"github.com/Laky-64/goswift/demangling/utils"
	"github.com/Laky-64/goswift/proxy"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

func getRawObjects(file *proxy.Context) ([]types.ObjectInfo, error) {
	exports, err := file.GetExports()
	if err != nil {
		return nil, err
	}
	packageNameRgx := regexp.MustCompile(`TelegramApi\.(Api|SecretApi[0-9]+)(\.functions)?\.(.*)`)
	parametersRgx := regexp.MustCompile(`TelegramApi\..*?\((.*?)\)`)
	objects := make(map[string]types.ObjectInfo)

	//cStrings, err := file.GetSwiftReflectionStrings()
	//cStrings, err := file.GetCStrings()
	//if err != nil {
	//	return nil, err
	//}
	//mapStrings := make(map[uint64]string)
	//for _, test := range cStrings {
	//	for t, addr := range test {
	//		if strings.HasPrefix(t, "pageBlock") {
	//			//fmt.Println(addr, t)
	//			mapStrings[addr] = t
	//		}
	//	}
	//	//if strings.HasPrefix(test, "pageBlock") {
	//	//	//fmt.Println(addr, test)
	//	//	//mapStrings[addr] = test
	//	//}
	//	//fmt.Println(test)
	//}
	//keys := slices.Collect(maps.Keys(mapStrings))
	//slices.Sort(keys)
	//slices.Reverse(keys)
	//for _, key := range keys {
	//	fmt.Println(key, mapStrings[key])
	//}

	for _, field := range exports {
		demangler, err := demangling.New([]byte(field.Name))
		if err != nil {
			continue
		}
		r, err := demangler.Result()
		if err != nil {
			continue
		}
		r = r.FirstChild()
		if r.Kind == demangling.StaticKind {
			r = r.FirstChild()
		}
		if r.Kind == demangling.FunctionKind {
			packageInfo := packageNameRgx.FindAllStringSubmatch(utils.ToString(r.FirstChild(), 0), -1)
			functionName := r.Children[1].Text
			isMethod := deepTextMatch(r, "Buffer", demangling.ReturnTypeKind, demangling.TupleKind, demangling.ClassKind)
			if len(packageInfo) > 0 && (functionName == "serialize" || isMethod) {
				var obj types.ObjectInfo
				var packageName strings.Builder
				packageName.WriteString(packageInfo[0][1])
				packageName.WriteString(".")
				if functionName == "serialize" {
					packageName.WriteString(packageInfo[0][3])
				} else {
					packageName.WriteString(packageInfo[0][3])
					packageName.WriteString(".")
					packageName.WriteString(functionName)
					obj.Parameters = parametersRgx.FindStringSubmatch(utils.ToString(r, utils.ModeNoSugar))[1]
				}
				obj.Name = strings.TrimPrefix(packageName.String(), fmt.Sprintf("%s.", packageInfo[0][1]))
				pInfo := strings.Split(obj.Name, ".")
				obj.Name = pInfo[len(pInfo)-1]
				obj.Package = strings.Join(pInfo[:len(pInfo)-1], ".")
				obj.Address = field.Address
				obj.IsMethod = isMethod
				obj.IsSecret = strings.HasPrefix(packageInfo[0][1], "Secret")
				if layerNumber := strings.TrimPrefix(packageInfo[0][1], "SecretApi"); len(layerNumber) > 0 {
					obj.Layer, _ = strconv.Atoi(layerNumber)
				}
				objects[packageName.String()] = obj
			}
		}
	}
	fields, err := file.GetSwiftFields()
	if err != nil {
		return nil, err
	}
	var TOBEDELETED []string
	for _, field := range fields {
		typeName := strings.TrimPrefix(field.Type, "TelegramApi.")
		if obj, ok := objects[typeName]; ok {
			schemeKind := strings.Split(typeName, ".")[0]
			delete(objects, typeName)
			mapStrings := make(map[uint64]string)
			for i, record := range field.Records {
				obj.Name = record.Name
				demangled, err := file.Demangle(record.MangledTypeNameOffset.GetAddress())
				if err != nil {
					return nil, err
				}
				obj.Parameters = strings.Trim(utils.ToString(demangled, utils.ModeNoSugar), "()")
				var kindName strings.Builder
				kindName.WriteString(schemeKind)
				kindName.WriteString(".")
				if len(obj.Package) > 0 {
					kindName.WriteString(obj.Package)
					kindName.WriteString(".")
				}
				obj.CaseOffset = i
				kindName.WriteString(record.Name)
				kindStringName := kindName.String()
				objects[kindStringName] = obj
				if typeName == "Api.PageBlock" {
					TOBEDELETED = append(TOBEDELETED, kindStringName)
					//fmt.Println(kindStringName, record.MangledTypeNameOffset.GetAddress(), obj.CaseOffset)
				}
			}

			keys := slices.Collect(maps.Keys(mapStrings))
			slices.Sort(keys)
			//slices.Reverse(keys)
			for _, key := range keys {
				fmt.Println(key, mapStrings[key])
			}
			if len(field.Records) > 20 && len(field.Records) < 130 {
				fmt.Println("FOUND BIG", typeName, len(field.Records))
			}
		}
	}
	compareUp := make(map[string]string)
	for _, obj := range environment.LocalStorage.PreviewLayer.MainApi.Constructors {
		compareUp[obj.Package()] = scheme.ParseConstructor(obj.Constructor())
	}
	objectsWithConstructors := make(map[string]types.ObjectInfo)
	for t, obj := range objects {
		instructions, err := asm.GetInstructions(file, obj.Address)
		if err != nil {
			return nil, err
		}
		//if t != "Api.inputPeerUserFromMessage" {
		//	continue
		//}
		//if t != "Api.account.savedRingtone" {
		//	continue
		//}
		//if t != "Api.updateBotBusinessConnect" {
		//	continue
		//}
		if t != "Api.pageBlockAnchor" {
			continue
		}
		if !obj.IsMethod {
			fmt.Println(t)
			asm.DebugAsm(instructions, TOBEDELETED)
			caseInstr, err := asm.ExtractFromSwitch(instructions, obj.CaseOffset)
			if err != nil && !errors.Is(err, asm.NoCasesError) {
				gologging.Fatal("Failed to extract from switch", t, err)
				continue
			} else if err == nil {
				instructions = caseInstr
			}
		}
		constructor, err := extractConstructor(instructions)
		if err != nil {
			return nil, err
		}
		var nameNoCollision strings.Builder
		if obj.IsSecret {
			nameNoCollision.WriteString("Secret")
			nameNoCollision.WriteString(fmt.Sprintf("%d", obj.Layer))
			nameNoCollision.WriteString(".")
		}
		nameNoCollision.WriteString(constructor)
		var tmpFullName strings.Builder
		tmpFullName.WriteString(obj.Package)
		if len(obj.Package) > 0 {
			tmpFullName.WriteString(".")
		}
		tmpFullName.WriteString(obj.Name)
		gologging.Debug("Found constructor", t, obj.Address, constructor, obj.CaseOffset)
		if !obj.IsSecret && compareUp[tmpFullName.String()] != constructor && obj.CaseOffset > 0 {
			gologging.Fatal("Constructor mismatch", t, compareUp[tmpFullName.String()], constructor, obj.CaseOffset)
		}
		objectsWithConstructors[nameNoCollision.String()] = obj
	}
	gologging.Debug("Objects with constructors", len(objectsWithConstructors))

	return slices.Collect(maps.Values(objects)), nil
}
