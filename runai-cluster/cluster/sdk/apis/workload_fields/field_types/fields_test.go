package field_types

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
	"testing"
)

const ProtocolHTTP = "http"

func TestPrimitives(t *testing.T) {

	strFldNonDefault := NewOptionalStringField("hello", "world")
	assert.Equal(t, "hello", strFldNonDefault.Deref(), "construct optional string - non default")

	strFldDefault := NewOptionalStringField("hello", "hello")
	assert.Equal(t, "", strFldDefault.Deref(), "construct optional string - default")

	strFldWithValue := NewStringField("aaa")
	assert.Equal(t, "aaa", strFldWithValue.Deref(), "deref str with value")
	assert.Equal(t, "aaa", strFldWithValue.DerefOrDefault("bbb"), "deref w/default str with value")

	strFldEmpty := NewStringField("")
	assert.Equal(t, "", strFldEmpty.Deref(), "deref empty str")
	assert.Equal(t, "", strFldEmpty.DerefOrDefault("bbb"), "deref w/default empty str")

	strFldNull := StringField{Value: nil}
	assert.Equal(t, "", strFldNull.Deref(), "deref null str")
	assert.Equal(t, "bbb", strFldNull.DerefOrDefault("bbb"), "deref w/default null str")

	boolWithValue := NewOptionalBoolField(ptr.To(true))
	assert.Equal(t, true, boolWithValue.Deref(), "deref bool/true")

	boolWithNull := NewOptionalBoolField(nil)
	assert.Equal(t, false, boolWithNull.Deref(), "deref bool/nil")

	intWithValue := NewOptionalIntField(ptr.To(35))
	assert.Equal(t, 35, intWithValue.Deref(), "deref int with value")
	assert.Equal(t, int32(35), intWithValue.Deref32(), "deref32 int with value")
	assert.Equal(t, int32(35), intWithValue.Deref32default(55), "deref32/value int with value")

	intWithNull := NewOptionalIntField(nil)
	assert.Equal(t, 0, intWithNull.Deref(), "deref int/null")
	assert.Equal(t, int32(0), intWithNull.Deref32(), "deref32 int/null")
	assert.Equal(t, int32(55), intWithNull.Deref32default(int32(55)), "deref32 with default int/null")

	var floatVal float64 = 32.0
	numberWithValue := NewOptionalNumberField(ptr.To(floatVal))
	assert.Equal(t, floatVal, numberWithValue.Deref(), "deref float with value")
	assert.Equal(t, "32", numberWithValue.DerefStr(), "deref/str float with value")

	var zeroFloat float64 = 0.0
	numberWithNull := NewOptionalNumberField(nil)
	assert.Equal(t, zeroFloat, numberWithNull.Deref(), "deref float with null")
	assert.Equal(t, "", numberWithNull.DerefStr(), "deref/str float with null")
}

func TestDictionary(t *testing.T) {

	envVarsMap := make(map[string]string)
	envVarsInR := map[string]DictionaryItemAndRules{}

	nullDict := DictionaryField{Value: nil}
	assert.Equal(t, envVarsInR, nullDict.Deref(), "deref null dict")

	nullDict.Put("USER", "tester")
	usersMap := nullDict.Deref()
	assert.Equal(t, "tester", ptr.Deref(usersMap["USER"].Value, ""), "put to nil dictionary")

	dict := NewDictionaryField(envVarsMap)
	assert.Equal(t, envVarsInR, dict.Deref(), "deref empty env vars")

	dict.Put("HOME", "/home")
	envVarsMap["HOME"] = "/home"
	envVarsInR["HOME"] = DictionaryItemAndRules{
		Value: ptr.To("/home"),
	}
	assert.Equal(t, envVarsInR, dict.Deref(), "deref HOME env var")

	newButSameDict := NewDictionaryField(envVarsMap)
	assert.Equal(t, envVarsInR, newButSameDict.Deref(), "deref same env vars")

}

func TestGitSync(t *testing.T) {

	gitSyncMap := make(map[string]GitSyncItemAndRules)
	gitSyncFld := NewGitSyncField(nil)
	assert.Equal(t, gitSyncMap, gitSyncFld.Deref(), "null git sync")

	gitSyncMap["mygit"] = GitSyncItemAndRules{
		Value: &GitSyncItemValue{
			Repository: ptr.To("myrepo"),
			Branch:     ptr.To("mybrancch"),
			Revision:   ptr.To("myrevision"),
			Username:   ptr.To("myuser"),
			Password:   ptr.To("mypass"),
			ReadOnly:   ptr.To(true),
			Image:      ptr.To("myimage"),
		},
	}
	assert.Equal(t, gitSyncMap, NewGitSyncField(gitSyncMap).Deref(), "single git sync")
}

func TestPorts(t *testing.T) {

	portsMap := make(map[string]PortItemAndRules)
	portsFld := NewPortsField(nil)
	assert.Equal(t, portsMap, portsFld.Deref(), "null ports")

	portsMap["myport"] = PortItemAndRules{
		Value: &PortItemValue{
			ContainerPort:   ptr.To(30000),
			HostPort:        ptr.To(40000),
			AutoGenHostPort: ptr.To(true),
			Protocol:        ptr.To(ProtocolHTTP),
		},
	}
	assert.Equal(t, portsMap, NewPortsField(portsMap).Deref(), "single port")
}

func TestPvcs(t *testing.T) {

	pvcsMap := make(map[string]PvcsItemAndRules)
	pvcsFld := NewPvcsField(nil)
	assert.Equal(t, pvcsMap, pvcsFld.Deref(), "null pvcs")

	pvcsMap["mypvc"] = PvcsItemAndRules{
		Value: &PvcsItemValue{
			StorageClass: ptr.To("mystorage"),
			Size:         ptr.To("mysize"),
			Path:         ptr.To("mypath"),
			ReadOnly:     ptr.To(true),
		},
	}
	assert.Equal(t, pvcsMap, NewPvcsField(pvcsMap).Deref(), "single pvc")
}

func TestVolumes(t *testing.T) {

	volumesMap := make(map[string]VolumeItemAndRules)
	volumesFld := NewVolumesField(nil)
	assert.Equal(t, volumesMap, volumesFld.Deref(), "null volumes")

	volumesMap["myvolume"] = VolumeItemAndRules{
		Value: &VolumeItemValue{
			SourcePath: ptr.To("mysrc"),
			TargetPath: ptr.To("mytarget"),
			ReadOnly:   ptr.To(true),
			NfsServer:  ptr.To("mynfs"),
		},
	}
	volumesMap["hisvolume"] = VolumeItemAndRules{
		Value: &VolumeItemValue{
			SourcePath: ptr.To("hissrc"),
			TargetPath: ptr.To("histarget"),
			ReadOnly:   nil,
			NfsServer:  ptr.To("hisnfs"),
		},
	}
	assert.Equal(t, volumesMap, NewVolumesField(volumesMap).Deref(), "two volumes")
}
