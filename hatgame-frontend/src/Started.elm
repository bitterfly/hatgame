module Started exposing (Data, ProcessState(..), addResultsAndSort)

import Containers.Game
import Containers.User


type ProcessState
    = StorytellerWaiting
    | StorytellerActive
    | NotStoryteller (Maybe Containers.User.User)
    | BetweenStages


type alias Data =
    { game : Containers.Game.Game
    , currentWord : Maybe String
    , partner : Maybe Containers.User.User
    , timer : Maybe Int
    , processState : ProcessState
    , results : List Containers.Game.Team
    }


addResults : List Containers.Game.Team -> List Containers.Game.Team -> Maybe (List Containers.Game.Team)
addResults first second =
    case ( first, second ) of
        ( [], [] ) ->
            Just []

        ( xs, [] ) ->
            Just xs

        ( x :: xs, y :: ys ) ->
            case addResults xs ys of
                Nothing ->
                    Nothing

                Just res ->
                    Just <|
                        { playerOne = x.playerOne
                        , playerTwo = x.playerTwo
                        , score = x.score + y.score
                        }
                            :: res

        ( _, _ ) ->
            Nothing


addResultsAndSort : List Containers.Game.Team -> List Containers.Game.Team -> List Containers.Game.Team
addResultsAndSort first second =
    case addResults first second of
        Nothing ->
            []

        Just res ->
            List.sortBy (\x -> x.score) res
