module Home.Http exposing (getCurrentUserToken, getStats, request)

import Containers.Statistics
import Containers.User
import Home
import Http
import Msg exposing (Msg)
import Url.Builder


request : { model | backend : String } -> { r | sessionToken : String } -> { gameId : Int } -> Cmd Msg
request { backend } { sessionToken } { gameId } =
    Http.request
        { method = "POST"
        , headers = [ Http.header "Authorization" ("bearer " ++ sessionToken) ]
        , url =
            Url.Builder.crossOrigin
                backend
                [ "game", "id", String.fromInt gameId ]
                []
        , body = Http.emptyBody
        , expect = Http.expectWhatever Msg.GameOk
        , timeout = Nothing
        , tracker = Nothing
        }


getStats : { model | backend : String } -> { r | sessionToken : String } -> Cmd Msg
getStats { backend } { sessionToken } =
    Http.request
        { method = "GET"
        , headers = [ Http.header "Authorization" ("bearer " ++ sessionToken) ]
        , url =
            Url.Builder.crossOrigin
                backend
                [ "stat" ]
                []
        , body = Http.emptyBody
        , expect = Http.expectJson Msg.GotStats Containers.Statistics.decode
        , timeout = Nothing
        , tracker = Nothing
        }


getCurrentUserToken : { model | backend : String } -> { r | sessionToken : String } -> Cmd Msg
getCurrentUserToken { backend } { sessionToken } =
    Http.request
        { method = "POST"
        , headers = [ Http.header "Authorization" ("bearer " ++ sessionToken) ]
        , url =
            Url.Builder.crossOrigin
                backend
                [ "user" ]
                []
        , body = Http.emptyBody
        , expect = Http.expectJson Msg.GotCurrentUserToken Containers.User.decode
        , timeout = Nothing
        , tracker = Nothing
        }
