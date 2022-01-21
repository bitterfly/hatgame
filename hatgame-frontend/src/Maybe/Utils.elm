module Maybe.Utils exposing (ChangeMaybeInt(..), fromString, toMaybeInt)


fromString : String -> Maybe String
fromString str =
    case String.length str of
        0 ->
            Nothing

        _ ->
            Just str


type ChangeMaybeInt
    = NoString
    | NotInt
    | ParsedInt Int


toMaybeInt : String -> ChangeMaybeInt
toMaybeInt str =
    case String.length str of
        0 ->
            NoString

        _ ->
            case String.toInt str of
                Nothing ->
                    NotInt

                Just n ->
                    ParsedInt n
