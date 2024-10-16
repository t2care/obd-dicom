[![test](https://github.com/t2care/obd-dicom/actions/workflows/pr.yml/badge.svg)](https://github.com/t2care/obd-dicom/actions/workflows/pr.yml)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=t2care_obd-dicom&metric=bugs)](https://sonarcloud.io/summary/new_code?id=t2care_obd-dicom)

# obd-dicom

One Byte Data DICOM Golang Library

## Install

```bash
go get -u github.com/t2care/obd-dicom
```

## Plugin JPEG/JPEG2000:

```bash
go build -tags "jpeg jpeg2000" ...
```

## CMD:

### WorklistSCU

```bash
go run cmd/obd-dicom/main.go  -cfindWorklist -calledae=SCP -host=x.x.x.x -port=y
```

### Modify Dicom file

```bash
go run cmd/obd-dicom/main.go -modify PatientName=abc,PatientAddress=123 -file samples/test.dcm
```

## Usage

### Load DICOM File

```golang
obj, err := media.NewDCMObjFromFile(fileName, &ParseOptions{SkipPixelData: true})
if err != nil {
  log.Panicln(err)
}
obj.DumpTags()
```

### Update string tag

```golang
obj, _ := media.NewDCMObjFromFile(fileName, &ParseOptions{SkipPixelData: true})
obj.WriteString(tags.PatientName, "new value")
obj.WriteToFile(fileName)
```

### Send C-Echo Request
```golang
scu := services.NewSCU(destination)
err := scu.EchoSCU(0)
if err != nil {
  log.Fatalln(err)
}
log.Println("CEcho was successful")
```

### Send C-Find Request
```golang
request := utils.DefaultCFindRequest()
scu := services.NewSCU(destination)
scu.SetOnCFindResult(func(result media.DcmObj) {
  log.Printf("Found study %s\n", result.GetString(tags.StudyInstanceUID))
  result.DumpTags()
})

count, status, err := scu.FindSCU(request, 0)
if err != nil {
  log.Fatalln(err)
}
```

### Send C-Store Request: Multiple files and Transcode are supported
```golang
scu := services.NewSCU(destination)
err := scu.StoreSCU([]string{fileName}, 0)  // By default ImplicitVRLittleEndian and JPEGLosslessSV1 will be proposed 
// err := scu.StoreSCU([]string{fileName}, 0, []string{transfersyntax.ExplicitVRLittleEndian.UID}) // Force transcoding to ExplicitVRLittleEndian
if err != nil {
  log.Fatalln(err)
}
```

### Send C-Move Request
```golang
request := utils.DefaultCMoveRequest(studyUID)

scu := services.NewSCU(destination)
_, err := scu.MoveSCU(destinationAE, request, 0)
if err != nil {
  log.Fatalln(err)
}
```

### Start SCP Server
```golang
scp := services.NewSCP(*port)

scp.OnAssociationRequest(func(request network.AAssociationRQ) bool {
  called := request.GetCalledAE()
  return *calledAE == called
})

scp.OnAssociationRelease(func(request network.AAssociationRQ) {
  request.GetID()
})

scp.OnCFindRequest(func(request network.AAssociationRQ, queryLevel string, query media.DcmObj) ([]media.DcmObj, uint16) {
  query.DumpTags()
  results := make([]media.DcmObj, 0)
  for i := 0; i < 10; i++ {
    results = append(results, utils.GenerateCFindRequest())
  }
  return results, dicomstatus.Success
})

scp.OnCMoveRequest(func(request network.AAssociationRQ, moveLevel string, query media.DcmObj) uint16 {
  query.DumpTags()
  return dicomstatus.Success
})

scp.OnCStoreRequest(func(request network.AAssociationRQ, data media.DcmObj) uint16 {
  log.Printf("INFO, C-Store recieved %s", data.GetString(tags.SOPInstanceUID))
  directory := filepath.Join(*datastore, data.GetString(tags.PatientID), data.GetString(tags.StudyInstanceUID), data.GetString(tags.SeriesInstanceUID))
  os.MkdirAll(directory, 0755)

  path := filepath.Join(directory, data.GetString(tags.SOPInstanceUID)+".dcm")

  // Lossless compression 
  if err := data.ChangeTransferSynx(transfersyntax.JPEGLosslessSV1); err != nil{
    log.Printf("ERROR: Compression %s : %s", path, err.Error())
  }
  err := data.WriteToFile(path)
  if err != nil {
    log.Printf("ERROR: There was an error saving %s : %s", path, err.Error())
  }
  return dicomstatus.Success
})

err := scp.Start()
if err != nil {
  log.Fatal(err)
}
```
