package main

import (
	"fmt"
	"math"

	lua "github.com/yuin/gopher-lua"
)

const AppType int = math.MaxInt

func (a *App) parseLua(luab []byte, appid int, name string) error {
	L := lua.NewState()
	defer L.Close()

	L.SetGlobal("addappid", L.NewFunction(func(l *lua.LState) int {
		// addappid(appid)
		// addappid(depotid, ?)
		// addappid(app/depotid, ?, "<key>")
		id := l.CheckInt(1)
		t := l.OptInt(2, AppType) // Depot type, we don't need this value
		key := l.OptString(3, "")

		if key == "" {
			if t == AppType {
				if AppendIntToSeq(a.AdditionalApps, id, name) {
					a.Out.Verbose("+ Lua: Added appid %d", id)
				}
			} else if AppendIntToSeq(a.AdditionalDepots, id, name) {
				a.Out.Verbose("+ Lua: Added depot %d (type %d)", id, t)
			}
		} else {
			SetMapKey(a.DecryptionKeys, id, key, name)
			if id != appid {
				if AppendIntToSeq(a.AdditionalDepots, id, name) {
					a.Out.Verbose("+ Lua: Added depot %d with key", id)
				}
			} else {
				a.Out.Verbose("+ Lua: Added appid key")
				AppendIntToSeq(a.AdditionalApps, appid, name)
			}
		}
		return 0
	}))
	L.SetGlobal("addtoken", L.NewFunction(func(l *lua.LState) int {
		// addtoken(appid, "<token>")
		app := l.CheckInt(1)
		token := l.CheckString(2)
		SetMapKey(a.AppTokens, app, token, name)
		a.Out.Verbose("+ Lua: Added apptoken %d", app)
		return 0
	}))
	L.SetGlobal("setManifestid", L.NewFunction(func(l *lua.LState) int {
		// setManifestid(depotid, "<gid>")
		depot := l.CheckInt(1)
		gid := l.CheckString(2)
		SetMapKey(a.ManifestIds, depot, gid, name)
		a.Out.Verbose("+ Lua: Pinned depot %d: %s", depot, gid)
		return 0
	}))

	mt := L.NewTable()
	L.SetField(mt, "__index", L.NewFunction(func(l *lua.LState) int {
		// If a unknown global is indexed, we blindly return
		// a ghost funtion which just logs its name and arguments,
		// to prevent unnecessary crashes.
		varName := l.CheckString(2)
		l.Push(l.NewFunction(func(l *lua.LState) int {
			top := l.GetTop()
			args := make([]string, 0, top)
			for i := 1; i <= top; i++ {
				args = append(args, l.Get(i).String())
			}
			a.Out.Verbose("- Lua: Ignored unknown call %q (args=%v)", varName, args)
			return 0
		}))
		return 1
	}))
	L.SetMetatable(L.GetGlobal("_G"), mt)

	if err := L.DoString(string(luab)); err != nil {
		return fmt.Errorf("Lua execution failed: %w", err)
	}
	if AppendIntToSeq(a.AdditionalApps, appid, name) {
		a.Out.Verbose("+ Lua: Added appid %d", appid)
	}
	return nil
}
