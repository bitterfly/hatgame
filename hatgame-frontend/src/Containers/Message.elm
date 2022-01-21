module Containers.Message exposing (Message(..), decode, decodeMsgString)

import Containers.Game
import Json.Decode exposing (Decoder)


type Message
    = Game Containers.Game.Game
    | Word String
    | Team Int
    | Story String
    | Tick Int
    | Started Int
    | Ended (List Containers.Game.Team)
    | Error String


decode : Decoder Message
decode =
    Json.Decode.andThen decodeMsgString (Json.Decode.field "Type" Json.Decode.string)


decodeMsgString : String -> Decoder Message
decodeMsgString str =
    case str of
        "game" ->
            Json.Decode.map Game
                (Json.Decode.field "Msg" Containers.Game.decode)

        "word" ->
            Json.Decode.map Word
                (Json.Decode.field "Msg" Json.Decode.string)

        "team" ->
            Json.Decode.map Team <| Json.Decode.field "Msg" Json.Decode.int

        "start" ->
            Json.Decode.map Started <| Json.Decode.field "Msg" Json.Decode.int

        "end" ->
            Json.Decode.map Ended <| Json.Decode.field "Msg" <| Json.Decode.list Containers.Game.decodeTeam

        "story" ->
            Json.Decode.map Story <| Json.Decode.field "Msg" Json.Decode.string

        "tick" ->
            Json.Decode.map Tick <| Json.Decode.field "Msg" Json.Decode.int

        "error" ->
            Json.Decode.map Error <| Json.Decode.field "Msg" Json.Decode.string

        x ->
            Json.Decode.fail <| "message not recognised " ++ x
