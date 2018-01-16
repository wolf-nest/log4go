package log4go

type FileWriter struct {
	level  int
}

func NewFileWriter(level int) *ConsoleWriter {
	var console = &ConsoleWriter{}
	console.level = level
	return console
}

func (this *FileWriter) WriteMessage(msg *LogMessage) {
	if msg == nil {
		return
	}
	if msg.level < this.level {
		return
	}
}

func (this *FileWriter) Close() {
}