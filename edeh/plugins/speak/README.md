# Speak Plugin for E:D Event Hub

## TTS Systems to Use

### espeak-ng
https://github.com/espeak-ng/espeak-ng/

### Peter's Text to Speech aka ptts.exe
http://jampal.sourceforge.net/ptts.html

### Windows Builtin TTS
As VBS script From [stackoverflow](https://stackoverflow.com/questions/1040655/ms-speech-from-command-line)

```
'say.vbs
set s = CreateObject("SAPI.SpVoice")
s.Speak Wscript.Arguments(0), 3
s.WaitUntilDone(1000)
```

call it as

```
cscript say.vbs "hello there"
```

Problem: One has (AKAIK) to guess the number for `s.WaitUntilDone`

With Powershell

```
AssemblyName System.Speech; (New-Object System.Speech.Synthesis.SpeechSynthesizer).Speak('hello');"
```

With mshta.exe

```
mshta vbscript:Execute("CreateObject(""SAPI.SpVoice"").Speak(""Hello"")(window.close)")
```
