package imp

type implementation struct {
	classUID string
	version  string
}

var imp *implementation

func SetDefaultImplementation() *implementation {
	imp = &implementation{
		classUID: "1.2.826.0.1.3680043.10.90.999",
		version:  "OBD-Dicom",
	}
	return imp
}

func SetImplementation(classUID string, version string) *implementation {
	imp = &implementation{
		classUID: classUID,
		version:  version,
	}
	return imp
}

func GetImpClassUID() string {
	if imp == nil {
		SetDefaultImplementation()
	}
	return imp.GetClassUID()
}

func GetImpVersion() string {
	if imp == nil {
		SetDefaultImplementation()
	}
	return imp.GetVersion()
}

func (i *implementation) GetClassUID() string {
	if i.classUID == "" {
		imp := SetDefaultImplementation()
		i.classUID = imp.GetClassUID()
		i.version = imp.GetVersion()
	}
	return i.classUID
}

func (i *implementation) GetVersion() string {
	if i.classUID == "" {
		imp := SetDefaultImplementation()
		i.classUID = imp.GetClassUID()
		i.version = imp.GetVersion()
	}
	return i.version
}
