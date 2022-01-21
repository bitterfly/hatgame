module Register.Http exposing (request)

import Containers.User
import Http
import Msg exposing (Msg)
import Url.Builder
import UserCredentials


request : { model | backend : String } -> { email : String, password : String, username : String } -> Cmd Msg
request { backend } { email, password, username } =
    Http.post
        { url =
            Url.Builder.crossOrigin
                backend
                [ "register" ]
                []
        , body =
            Http.jsonBody <|
                Containers.User.encodeRegister
                    { email = email
                    , password = password
                    , username = username
                    }
        , expect = Http.expectWhatever Msg.CheckRegister
        }
