module Started.View exposing (html)

import Containers.User
import Containers.Game
import Generic.Utils
import Html exposing (Html, button, div, h1, h5, h3, label, text)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)
import Msg exposing (Msg)
import Progress.Ring
import Started


html : Maybe Containers.User.WithToken -> Maybe String -> Started.Data -> List (Html Msg)
html tokenUser err startedData =
    [ div []
        [ div [ class "header-home" ]
            [ label [ class "user-label" ] [ text <| Containers.User.maybeUsernameWithToken tokenUser ]
            , text "you're playing with"
            , label
                [ class "user-label" ]
                [ text <| Containers.User.maybeUsername startedData.partner ]
            ]
        , div [ class "container" ]
            [ div [ class "row" ]
                [ div
                    [ classList
                        [ ( "shift-3", True )
                        , ( "has-error", err /= Nothing )
                        ]
                    ]
                    [ 
                        div
                            [ style "display" "flex"
                            , style
                                "flex-direction"
                                "column"
                            , style "justify-content" "space-around"
                            , style "align-items" "center"
                            ]
                            [ 
                                (if gameNotActive startedData.processState
                                then 
                                    div[][]
                                else
                                    timerView startedData)
                            , case startedData.processState of
                                Started.NotStoryteller user ->
                                    otherView startedData user

                                Started.StorytellerWaiting ->
                                    waitingView startedData

                                Started.StorytellerActive ->
                                    activeView startedData

                                Started.BetweenStages ishost ->
                                    betweenStagesView startedData ishost

                                Started.GameEnded ->
                                    endView startedData tokenUser
                            ]
                    ]
                ]
            ]
        ]
    ]

gameNotActive : Started.ProcessState -> Bool
gameNotActive ps =
    case ps of
    Started.BetweenStages _  -> True
    Started.GameEnded   -> True
    _ -> False


timerView : Started.Data -> Html Msg
timerView startedData =
    div[][
    div [ class "spacing-both" ] []
    ,Progress.Ring.view
        { color =
            Generic.Utils.makeColor
                (toFloat (Maybe.withDefault startedData.game.timer startedData.timer)
                    / toFloat startedData.game.timer
                )
        , progress =
            toFloat
                (Maybe.withDefault startedData.game.timer startedData.timer)
                / toFloat startedData.game.timer
        , stroke = 30
        , radius = 100
        }
    ,div [ class "spacing-both" ] []
    ]

otherView : Started.Data -> Maybe Containers.User.User -> Html Msg
otherView startedData user =
    div []
        [ h5 []
            [ text <|
                case user of
                    Nothing ->
                        "Watiing for storyteller."

                    Just u ->
                        case startedData.partner of
                            Nothing ->
                                ""

                            Just p ->
                                if u.id == p.id then
                                    Containers.User.show u ++ "'s turn. You're guessing."

                                else
                                    Containers.User.show u ++ "'s turn. Shhh, just listen!"
            ]
        ]


waitingView : Started.Data -> Html Msg
waitingView _ =
    div
        [ style "display" "flex"
        , style "flex-direction" "column"
        , style "justify-content" "center"
        , style "align-items" "center"
        ]
        [ button
            [ class "btn-primary"
            , onClick <| Msg.SendReadyStoryteller
            ]
            [ text "Start" ]
        ]


activeView : Started.Data -> Html Msg
activeView startedData =
    div
        [ style "display" "flex"
        , style "flex-direction" "column"
        , style "justify-content" "center"
        , style "align-items" "center"
        ]
        [ h1 [] [ text <| Maybe.withDefault "" startedData.currentWord ]
        , button
            [ class "btn-primary"
            , case startedData.currentWord of
                Nothing ->
                    disabled True

                Just word ->
                    onClick <| Msg.SendGuessed word
            ]
            [ text "Yep!" ]
        ]

endView : Started.Data ->  Maybe Containers.User.WithToken -> Html Msg
endView startedData tokenUser =
    div [ class "container" ]
        [ div [ class "row" ]
            [ div
                [ classList
                    [ ( "shift-3", True )
                    ]
                ]
                [ div [ class "spacing-both" ]
                    []
                ,(case tokenUser of
                Nothing ->
                    div[][]
                Just user ->
                    h3 [ style "text-align" "center" ] [ text <| Containers.Game.showResult 
                    <| Containers.Game.result user.user startedData.results ])
                ,showResults
                    startedData.game.players
                    startedData.results
                ,button
                    [ class "btn-primary"
                    , onClick <|
                        Msg.End
                    ]
                    [ text "End" ]
                
            ]
        ]
    ]


betweenStagesView : Started.Data -> Bool -> Html Msg
betweenStagesView startedData ishost =
    div [ class "container" ]
        [ div [ class "row" ]
            [ div
                [ classList
                    [ ( "shift-3", True )
                    ]
                ]
                [ div [ class "spacing-both" ]
                    []
                ,h3 [ style "text-align" "center" ] [ text "Scores so far"]
                ,showResults
                    startedData.game.players
                    startedData.results
                ,(if ishost
                then button
                    [ class "btn-primary"
                    , onClick <|
                        Msg.End
                    ]
                    [ text "Start" ]
                else div[][])
            ]
        ]
    ]

showResults : List Containers.User.User -> List Containers.Game.Team -> Html Msg
showResults players teams =
    div []
        [
        div [ class "spacing-both" ] []
        , div
            [ class "display-window"
            , style "display" "flex"
            , style
                "flex-direction"
                "column"
            , style "justify-content" "space-around"
            ]
          <|
            List.concat [ List.map (showResult players) teams ]
        ]


showResult : List Containers.User.User -> Containers.Game.Team -> Html msg
showResult players team =
    div
        [ style "display" "flex"
        , style "justify-content" "space-around"
        ]
        [ div []
            [ text <| Maybe.withDefault "" (Containers.Game.getUsername players team.playerOne)
            ]
        , div [] [ text <| Maybe.withDefault "" (Containers.Game.getUsername players team.playerTwo) ]
        , div [] [ text <| String.fromInt team.score ]
        ]
