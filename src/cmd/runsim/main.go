package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/langs"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/runsim"
	"github.com/quasilyte/roboden-game/scenes/staging"
	"github.com/quasilyte/roboden-game/serverapi"
)

func main() {
	timeoutFlag := flag.Int("timeout", 30, "simulation timeout in seconds")
	debugFlag := flag.Bool("debug", false, "whether to enable debug logs")
	trustFlag := flag.Bool("trust", false, "whether to allow 0 levelgen checksums")
	flag.Parse()

	replayDataBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	var replayData serverapi.GameReplay
	if err := json.Unmarshal(replayDataBytes, &replayData); err != nil {
		panic(err)
	}

	if replayData.LevelGenChecksum == 0 && !*trustFlag {
		panic(errors.New("replay has a zero levelgen checksum"))
	}

	config := gamedata.MakeLevelConfig(gamedata.ExecuteSimulation, replayData.Config)
	ctx := ge.NewContext(ge.ContextConfig{
		Mute:       true,
		FixedDelta: true,
	})
	ctx.Loader.OpenAssetFunc = assets.MakeOpenAssetFunc(ctx, "")
	ctx.Dict = langs.NewDictionary("en", 2)

	runsim.PrepareAssets(ctx)

	state := runsim.NewState(ctx)
	state.Persistent.Settings.DebugLogs = *debugFlag

	config.Finalize()

	controller := staging.NewController(state, config, nil)
	controller.SetReplayActions(replayData.Actions)
	simResult, err := runsim.Run(state, replayData.LevelGenChecksum, *timeoutFlag, controller)
	if err != nil {
		panic(err)
	}

	encodedResult, err := json.Marshal(simResult)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(encodedResult))
}
