'say.vbs (cscript say.vbs "hello there")
set s = CreateObject("SAPI.SpVoice")
s.Speak Wscript.Arguments(0), 3
s.WaitUntilDone(1000)
