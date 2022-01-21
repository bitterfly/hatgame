module Change.Http exposing (request)

import Containers.User
import Http
import Msg exposing (Msg)
import Url.Builder


request : { model | backend : String } -> { r | sessionToken : String } -> { email : String, password : String, username : String } -> Cmd Msg
request { backend } { sessionToken } { email, password, username } =
    Http.request
        { method = "POST"
        , headers = [ Http.header "Authorization" ("bearer " ++ sessionToken) ]
        , url =
            Url.Builder.crossOrigin
                backend
                [ "user", "change" ]
                []
        , body =
            Http.jsonBody <|
                Containers.User.encodeRegister
                    { email = email
                    , password = password
                    , username = username
                    }
        , expect = Http.expectWhatever Msg.ToHomePage
        , timeout = Nothing
        , tracker = Nothing
        }
