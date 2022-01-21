module Login.Http exposing (request)

import Containers.User
import Http
import Msg exposing (Msg)
import Url.Builder


request : { model | backend : String } -> { email : String, password : String } -> Cmd Msg
request { backend } { email, password } =
    Http.post
        { url =
            Url.Builder.crossOrigin
                backend
                [ "login" ]
                []
        , body = Http.jsonBody <| Containers.User.encodeLogin { email = email, password = password }
        , expect = Http.expectJson Msg.GotUserToken Containers.User.decode
        }
