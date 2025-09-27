# Color Rush Survival - Complete Game Rules & Specifications

## Game Overview
Color Rush Survival is a fast-paced 2D multiplayer elimination game where players must navigate to correct colored tiles before time expires. Last player(s) standing wins.

## Core Game Parameters

### Map Configuration
- **Grid Size:** 20x15 blocks (300 total blocks)
- **Block Types:** 16 colors initially
- **Block Distribution:** Equal distribution of each color (18-19 blocks per color)
- **Border:** Void space surrounding the playable area

### Player Specifications
- **Minimum Players:** 4
- **Maximum Players:** 16
- **Spawn Location:** Random valid colored block
- **Movement Speed:** 4 blocks per second

## Game Phases & Timing

### Phase 1: Lobby/Waiting
- **Duration:** Until minimum players join (4)
- **Player Actions:** Join game, set username
- **Server Actions:** Accept connections, validate usernames

### Phase 2: Game Start Preparation
- **Duration:** 5 seconds
- **Player Actions:** Free movement on map
- **Server Actions:** Initialize game state, broadcast countdown

### Phase 3: Round Cycle
Each round consists of 4 sub-phases:

#### Sub-Phase A: Color Call (1 second)
- Server selects target color(s)
- Broadcast color to all players
- Start rush phase timer

#### Sub-Phase B: Rush Phase (Variable duration)
- Players move to target color blocks
- Monitor player positions
- Count down timer

#### Sub-Phase C: Elimination Check (0.5 seconds)
- Verify player positions
- Eliminate players not on correct color block
- Update scores

#### Sub-Phase D: Round Transition (1 second)
- Remove eliminated players
- Apply map changes if applicable
- Prepare next round

## Timing Progression

### Rush Phase Duration by Round:
```
Rounds 1-3:   4.0 seconds
Rounds 4-6:   3.5 seconds
Rounds 7-9:   3.0 seconds
Rounds 10-12: 2.5 seconds
Rounds 13-15: 2.0 seconds
Rounds 16-18: 1.8 seconds
Rounds 19-21: 1.5 seconds
Rounds 22+:   1.2 seconds
```

## Player Mechanics

### Position Validation
- **Valid Position:** Player center overlaps with target color block
- **Boundary Cases:** If player is on multiple blocks, any one being correct color counts
- **Update Frequency:** Position checked 10 times per second


### Elimination Rules
- **Position Check:** Performed at exact end of rush phase
- **Network Lag:** 100ms grace period for position updates
- **Disconnection:** Player eliminated if disconnected during rush phase
- **Spectator Mode:** Eliminated players can watch remaining game

## Scoring System

### Base Scoring:
- **Survival Points:** 10 points per round survived
- **Elimination Bonus:** 5 points × (total players - current placement)
- **Speed Bonus:** +2 points if reached color with >1 second remaining
- **Perfect Bonus:** +50 points if reached color with >2 seconds remaining

### Special Bonuses:
- **Final Winner:** +100 points for last player standing
- **Endurance Bonus:** +200 points for surviving Round 25

### Streak Bonuses:
- **3 Round Streak:** +30 points
- **5 Round Streak:** +75 points
- **10 Round Streak:** +200 points

## Victory Conditions

### Primary Victory:
- **Last Player Standing:** Solo winner
- **Multiple Survivors at Round 25:** Shared victory
- **All Players Eliminated:** No winner (rare edge case)

### Secondary Victory (if multiple survivors remain):
- **Highest Score:** Tiebreaker #1
- **Most Rounds Survived:** Tiebreaker #2
- **Fastest Average Response Time:** Tiebreaker #3

## Technical Specifications

### Server Tick Rate:
- **Position Updates:** 10 Hz (every 100ms)
- **Timer Updates:** 20 Hz (every 50ms)
- **Game State Broadcasts:** Variable based on phase

### Network Protocol:
- **Player Movement:** Real-time position updates
- **Color Announcements:** Reliable delivery required
- **Timer Sync:** Server authoritative
- **Lag Compensation:** 100ms buffer for position validation

### Data Persistence:
- **Game Results:** Store final scores and placements
- **Player Stats:** Track games played, wins, best scores
- **Leaderboards:** Daily, weekly, all-time rankings

## Game State Management

### State Transitions:
```
WAITING → PREPARATION → ROUND_START → COLOR_CALL → RUSH_PHASE → ELIMINATION → 
[ROUND_START] (loop) → GAME_FINISHED
```

### Error Handling:
- **Player Disconnect:** Immediate elimination
- **Server Lag:** Extend rush phase by lag duration
- **Invalid Moves:** Reject and maintain last valid position

### Reconnection Policy:
- **Grace Period:** 10 seconds during preparation phase
- **Mid-Round:** No reconnection allowed
- **Spectator Return:** Can rejoin as spectator only

## Anti-Cheat Measures

### Movement Validation:
- **Speed Limits:** Enforce maximum movement speed
- **Teleportation Check:** Validate movement distances
- **Boundary Enforcement:** Prevent out-of-bounds movement

### Position Verification:
- **Server Authority:** Server has final say on player positions
- **Timestamp Validation:** Check for outdated position updates
- **Rubber Banding:** Correct client position if desync detected

## Configuration Parameters

### Adjustable Settings:
- Map size (width, height)
- Number of colors
- Round timing progression
- Score multipliers
- Maximum players per game

## Event Logging

### Required Logs:
- Player join/leave events
- Round start/end with timings
- Color announcements and player responses
- Elimination events with reasons
- Score updates and final results
- Error conditions and recoveries

This specification provides a complete foundation for implementing the Color Rush Survival game backend.