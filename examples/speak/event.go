package speak

import (
	"fmt"
	"strings"
	"text/template"

	"git.fractalqb.de/fractalqb/daq"
	"git.fractalqb.de/fractalqb/sllm/v3"
	"github.com/mitchellh/mapstructure"
)

type evtHandler interface {
	configure(evt string, cfg any) error
	message(stat *status, evt any) (txt string, flags []string)
}

var handlers = map[string]evtHandler{
	"ReceiveText":         new(receiveText),
	"FSSSignalDiscovered": new(fssSignalDiscovered),
}

type EventMsg struct {
	Text     string
	Template string
	tmpl     *template.Template
}

func (em *EventMsg) configure(evt string) (err error) {
	if em.Template != "" {
		em.tmpl, err = template.New(evt).Parse(em.Template)
		if err != nil {
			return err
		}
		log.Info("`event` `template`", `event`, evt, `template`, em.Template)
	} else if em.Text == "" {
		return sllm.ErrorIdx("empty message for `event`", evt)
	} else {
		log.Info("`event` `text`: %s", `event`, evt, `text`, em.Text)
	}
	return nil
}

func (em *EventMsg) output(jevt any) string {
	if em.tmpl != nil {
		var sb strings.Builder
		err := em.tmpl.Execute(&sb, jevt)
		if err != nil {
			return err.Error()
		}
		return sb.String()
	}
	return em.Text
}

// defaultEvent will be used on every configured Events element that has no
// pre-registered handler
type defaultEvent struct {
	Flags []string `json:",omitempty"`
	Speak EventMsg
}

func (evt *defaultEvent) configure(e string, cfg any) error {
	if err := mapstructure.Decode(cfg, evt); err != nil {
		return err
	}
	return evt.Speak.configure(e)
}

func (evt *defaultEvent) message(stat *status, jevt any) (string, []string) {
	return evt.Speak.output(jevt), evt.Flags
}

type receiveText struct {
	defaultEvent
	Channels map[string]*defaultEvent
}

func (evt *receiveText) configure(e string, cfg any) error {
	if err := mapstructure.Decode(cfg, evt); err != nil {
		return fmt.Errorf("%s config: %w", e, err)
	}
	if evt.Speak.Text == "" && evt.Speak.Template == "" {
		if len(evt.Channels) == 0 {
			return fmt.Errorf("no outpur for %s", e)
		}
	} else if err := evt.defaultEvent.Speak.configure(e); err != nil {
		return fmt.Errorf("%s config: %w", e, err)
	}
	for c, ce := range evt.Channels {
		if err := ce.Speak.configure(e + "-" + c); err != nil {
			return fmt.Errorf("%s-%s config: %w", e, c, err)
		}
	}
	return nil
}

func (evt *receiveText) message(stat *status, jevt any) (string, []string) {
	chn := daq.Val(daq.AsString, "")(daq.Get(jevt, "Channel"))
	if chnCfg, ok := evt.Channels[chn]; ok {
		return chnCfg.Speak.output(jevt), chnCfg.Flags
	}
	return evt.Speak.output(jevt), evt.Flags
}

type fssSignalDiscovered struct {
	defaultEvent
}

func (evt *fssSignalDiscovered) message(stat *status, jevt any) (string, []string) {
	if stat.population > 0 {
		return "", nil
	}
	if stn := daq.Val(daq.AsBool, false)(daq.Get(jevt, "IsStation")); stn {
		return "", nil
	}
	return evt.Speak.output(jevt), evt.Flags
}
