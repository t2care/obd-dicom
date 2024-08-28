package imp

type Implementation struct {
	classUID string
	version  string
}

var imp *Implementation

func SetDefaultImplementation() *Implementation {
	imp = &Implementation{
		classUID: "1.2.826.0.1.3680043.10.90.999",
		version:  "OBD-Dicom",
	}
	return imp
}

func SetImplementation(classUID string, version string) *Implementation {
	imp = &Implementation{
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

func (i *Implementation) GetClassUID() string {
	if i.classUID == "" {
		imp := SetDefaultImplementation()
		i.classUID = imp.GetClassUID()
		i.version = imp.GetVersion()
	}
	return i.classUID
}

func (i *Implementation) GetVersion() string {
	if i.classUID == "" {
		imp := SetDefaultImplementation()
		i.classUID = imp.GetClassUID()
		i.version = imp.GetVersion()
	}
	return i.version
}
