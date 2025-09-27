0. the player should create a new game if the user do not have a room 

1. pre-game phase 
    - Data: 
    - Behavior:
        - In this phrase, the players can join the game.
            - When we got 2+ players, enter the next phase. 
2. pre-game-countdown phase
    - Data:
        - countdown: from 10
    - Behavior:
        - This phrase is meant for more players to join the game.
        - Every second during this phrase will decrease the countdown by 1
        - When the countdown reaches 0, enter next phrase
3. in-game phase
    - Data:
        - countdown: number
        - current_countdown: number
        - target_color: WoolColor
        - dead: array[player]
    - Behavior:
        - This is where the game actually runs
        - We will run this phrase in a **Round Loop**, which persists of the following steps:
            1. Generate a new map
            2. Determine a `target_color`, and set `this.target_color` to it
            3. Set the `current_countdown` to `countdown`, then decrease the `current_countdown` b y 1 every second.
            4. When the clock reaches 0, remove all the blocks other than `target_color`
            5. Check the block under every player, if their block underneath is air (16), elimate them, and add the elimated player into `dead` in order.
            6. Update the `countdown` based on the amount of rounds, make it lower every round.
            7. Go back to the step 1, repeat until all the players are elimated. then go to the next phase
    - Note:
        - the `current_countdown` is universal, both for the rush phrase and the rest before next wave. so the frontend and backend can just update the countdown on the same value. 
4. settlement phase
    - Behavior:
        - Just return the result, like dead player, nothing else.
    
            
# The Data a Game should hold
- It's map
- It's players and their states
- The phase-specific datas described above
