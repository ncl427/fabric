package tests

import (
	"github.com/openziti/channel"
	"github.com/openziti/fabric/controller"
)

func (ctx *TestContext) NewControlChannelListener() channel.UnderlayListener {
	config, err := controller.LoadConfig(ControllerConfFile)
	ctx.Req.NoError(err)

	versionHeader, err := VersionProviderTest{}.EncoderDecoder().Encode(VersionProviderTest{}.AsVersionInfo())
	ctx.Req.NoError(err)
	headers := map[int32][]byte{
		channel.HelloVersionHeader: versionHeader,
	}

	ctrlChannelListenerConfig := channel.ListenerConfig{
		ConnectOptions: config.Ctrl.Options.ConnectOptions,
		Headers:        headers,
	}
	ctrlListener := channel.NewClassicListener(config.Id, config.Ctrl.Listener, ctrlChannelListenerConfig)
	ctx.Req.NoError(ctrlListener.Listen())
	return ctrlListener
}
