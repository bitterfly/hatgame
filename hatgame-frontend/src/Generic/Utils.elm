module Generic.Utils exposing (errorToString, makeColor)

import Http


errorToString : Http.Error -> String
errorToString error =
    case error of
        Http.BadUrl url ->
            "The URL " ++ url ++ " was invalid"

        Http.Timeout ->
            "Unable to reach the server, try again"

        Http.NetworkError ->
            "Unable to reach the server, check your network connection"

        Http.BadStatus n ->
            case n of
                500 ->
                    "The server had a problem, try again later"

                400 ->
                    "Verify your information and try again"

                401 ->
                    "Unauthorized"

                _ ->
                    "Unknown error"

        Http.BadBody errorMessage ->
            errorMessage


makeColor : Float -> String
makeColor f =
    "rgb("
        ++ (String.join "," <|
                List.map String.fromFloat
                    [ (115 - 193) * f + 194
                    , (168 - 56) * f + 56
                    , 50
                    ]
           )
        ++ ")"
