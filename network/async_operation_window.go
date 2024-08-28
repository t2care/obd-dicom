package network

import "github.com/one-byte-data/obd-dicom/media"

type AsyncOperationWindow struct {
	ItemType                     byte //0x53
	Reserved1                    byte
	Length                       uint16
	MaxNumberOperationsInvoked   uint16
	MaxNumberOperationsPerformed uint16
}

// NewAsyncOperationWindow - NewAsyncOperationWindow
func NewAsyncOperationWindow() *AsyncOperationWindow {
	return &AsyncOperationWindow{
		ItemType: 0x53,
	}
}

func (async *AsyncOperationWindow) GetMaxNumberOperationsInvoked() uint16 {
	return async.MaxNumberOperationsInvoked
}

func (async *AsyncOperationWindow) GetMaxNumberOperationsPerformed() uint16 {
	return async.MaxNumberOperationsPerformed
}

func (async *AsyncOperationWindow) Size() uint16 {
	return async.Length + 4
}

func (async *AsyncOperationWindow) Read(ms *media.MemoryStream) (err error) {
	if async.ItemType, err = ms.GetByte(); err != nil {
		return err
	}
	return async.ReadDynamic(ms)
}

func (async *AsyncOperationWindow) ReadDynamic(ms *media.MemoryStream) (err error) {
	if async.Reserved1, err = ms.GetByte(); err != nil {
		return err
	}
	if async.Length, err = ms.GetUint16(); err != nil {
		return err
	}
	if async.MaxNumberOperationsInvoked, err = ms.GetUint16(); err != nil {
		return err
	}
	if async.MaxNumberOperationsPerformed, err = ms.GetUint16(); err != nil {
		return err
	}
	return
}
