/*
	Copyright NetFoundry, Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package handler_ctrl

import (
	"github.com/golang/protobuf/proto"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel"
	"github.com/openziti/fabric/controller/network"
	"github.com/openziti/fabric/controller/xctrl"
	"github.com/openziti/fabric/pb/ctrl_pb"
	"github.com/openziti/foundation/transport"
	"github.com/pkg/errors"
)

type CtrlAccepter struct {
	network      *network.Network
	xctrls       []xctrl.Xctrl
	listener     channel.UnderlayListener
	options      *channel.Options
	traceHandler *channel.TraceHandler
}

func NewCtrlAccepter(network *network.Network,
	xctrls []xctrl.Xctrl,
	listener channel.UnderlayListener,
	options *channel.Options,
	traceHandler *channel.TraceHandler) *CtrlAccepter {
	return &CtrlAccepter{
		network:      network,
		xctrls:       xctrls,
		listener:     listener,
		options:      options,
		traceHandler: traceHandler,
	}
}

func (self *CtrlAccepter) Run() {
	log := pfxlog.Logger()
	log.Info("started")
	defer log.Warn("exited")

	for {
		ch, err := channel.NewChannel("ctrl", self.listener, channel.BindHandlerF(self.Bind), self.options)
		if err != nil {
			log.WithError(err).Error("error accepting control channel connection")
			if err.Error() == "closed" {
				return
			}
		} else if r, err := self.network.GetRouter(ch.Id().Token); err == nil {
			go self.network.ConnectRouter(r)
		} else {
			log.WithError(err).Error("error getting router for control channel")
		}
	}
}

// TODO: Tod - This is the controller side of the world
func (self *CtrlAccepter) Bind(binding channel.Binding) error {
	binding.GetChannel().SetLogicalName(binding.GetChannel().Id().Token)
	ch := binding.GetChannel()

	log := pfxlog.Logger().WithField("routerId", ch.Id().Token)

	if r, err := self.network.GetRouter(ch.Id().Token); err == nil {
		if ch.Underlay().Headers() != nil {
			if versionValue, found := ch.Underlay().Headers()[channel.HelloVersionHeader]; found {
				if versionInfo, err := self.network.VersionProvider.EncoderDecoder().Decode(versionValue); err == nil {
					r.VersionInfo = versionInfo
				} else {
					return errors.Wrap(err, "could not parse version info from router hello, closing router connection")
				}
			} else {
				return errors.New("no version info header, closing router connection")
			}
			r.Listeners = nil
			if val, found := ch.Underlay().Headers()[int32(ctrl_pb.ContentType_ListenersHeader)]; found {
				log.Debug("router reported listeners using listeners header")
				listeners := &ctrl_pb.Listeners{}
				if err := proto.Unmarshal(val, listeners); err != nil {
					log.WithError(err).Error("unable to unmarshall listeners value")
				} else {
					for _, listener := range listeners.Listeners {
						var group = listener.GetGroup()
						if "" == group {
							group = "default"
						}
						log.WithField("address", listener.GetAddress()).WithField("type", listener.GetProtocol()).WithField("group", group).WithField("costTags", listener.GetCostTags()).Debug("router listener")
						r.AddLinkListener(listener.GetAddress(), listener.GetProtocol(), listener.GetGroup(), listener.GetCostTags())
					}
				}
			} else if listenerValue, found := ch.Underlay().Headers()[channel.HelloRouterAdvertisementsHeader]; found {
				log.Debug("router reported listener using advertisement header")
				addr := string(listenerValue)
				linkProtocol := "tls"

				if addr, _ := transport.ParseAddress(addr); addr != nil {
					linkProtocol = addr.Type()
				}
				r.AddLinkListener(addr, linkProtocol, "default", nil)
			} else {
				log.Warn("no advertised listeners")
			}

			r.Dialers = nil
			if val, found := ch.Underlay().Headers()[int32(ctrl_pb.ContentType_DialersHeader)]; found {
				log.Debug("router reported dialers using dialers header")
				dialers := &ctrl_pb.Dialers{}
				if err := proto.Unmarshal(val, dialers); err != nil {
					log.WithError(err).Error("unable to unmarshall dialers value")
				} else {
					for _, dialer := range dialers.Dialers {
						log.WithField("localBinding", dialer.GetLocalBinding()).WithField("group", dialer.GetGroup()).Debug("router dialer")
						r.AddLinkDialer(dialer.GetLocalBinding(), dialer.GetGroup())
					}
				}
			}
		} else {
			log.Warn("no attributes provided")
		}

		r.Control = ch
		if err := binding.Bind(newBindHandler(r, self.network, self.xctrls)); err != nil {
			return errors.Wrap(err, "error binding router")
		}

		if self.traceHandler != nil {
			binding.AddPeekHandler(self.traceHandler)
		}

		log.Infof("accepted new router connection [r/%s]", r.Id)
	}
	return nil
}
