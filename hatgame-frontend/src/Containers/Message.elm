module Containers.Message exposing (MessageReceived(..), MessageSend(..), decode, decodeMsgString, encodeMsgSend)

import Containers.Game
import Json.Decode exposing (Decoder)
import Json.Encode


type MessageReceived
    = Game Containers.Game.Game
    | ReceiveAddWord String
    | Team Int
    | Story String
    | Tick Int
    | GuessPhaseStart Int
    | Ended (List Containers.Game.Team)
    | Error String
    | ReadyToStart
    | WordPhaseStart


type MessageSend
    = ReadyStoryteller
    | RequestToStart
    | SendQuitLobby
    | Guess String
    | SendAddWord String


encodeMsgSend : MessageSend -> String
encodeMsgSend msg =
    Json.Encode.encode 0 <|
        case msg of
            ReadyStoryteller ->
                Json.Encode.object [ ( "type", Json.Encode.string "ready_storyteller" ) ]

            RequestToStart ->
                Json.Encode.object
                    [ ( "type", Json.Encode.string "request_to_start" ) ]

            Guess word ->
                Json.Encode.object
                    [ ( "type", Json.Encode.string "guess" )
                    , ( "msg", Json.Encode.string word )
                    ]

            SendAddWord word ->
                Json.Encode.object
                    [ ( "type", Json.Encode.string "add_word" )
                    , ( "msg", Json.Encode.string word )
                    ]

            SendQuitLobby ->
                Json.Encode.object
                    [ ( "type", Json.Encode.string "quit_lobby" ) ]


decode : Decoder MessageReceived
decode =
    Json.Decode.andThen decodeMsgString (Json.Decode.field "Type" Json.Decode.string)


decodeMsgString : String -> Decoder MessageReceived
decodeMsgString str =
    case str of
        "game" ->
            Json.Decode.map Game
                (Json.Decode.field "Msg" Containers.Game.decode)

        "add_word" ->
            Json.Decode.map ReceiveAddWord
                (Json.Decode.field "Msg" Json.Decode.string)

        "team" ->
            Json.Decode.map Team <| Json.Decode.field "Msg" Json.Decode.int

        "guess_phase_start" ->
            Json.Decode.map GuessPhaseStart <| Json.Decode.field "Msg" Json.Decode.int

        "end" ->
            Json.Decode.map Ended <| Json.Decode.field "Msg" <| Json.Decode.list Containers.Game.decodeTeam

        "story" ->
            Json.Decode.map Story <| Json.Decode.field "Msg" Json.Decode.string

        "tick" ->
            Json.Decode.map Tick <| Json.Decode.field "Msg" Json.Decode.int

        "error" ->
            Json.Decode.map Error <| Json.Decode.field "Msg" Json.Decode.string

        "ready_to_start" ->
            Json.Decode.succeed ReadyToStart

        "word_phase_start" ->
            Json.Decode.succeed WordPhaseStart

        x ->
            Json.Decode.fail <| "message not recognised " ++ x
