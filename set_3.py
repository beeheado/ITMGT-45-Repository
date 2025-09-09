'''Programming Set 3

This assignment will develop your ability to manipulate data.
'''

def relationship_status(from_member, to_member, social_graph):
    '''Relationship Status.

    Let us pretend that you are building a new app.
    Your app supports social media functionality, which means that users can have
    relationships with other users.

    There are two guidelines for describing relationships on this social media app:
    1. Any user can follow any other user.
    2. If two users follow each other, they are considered friends.

    This function describes the relationship that two users have with each other.

    Please see "set_3_sample_data.py" for sample data. The social graph
    will adhere to the same pattern.

    Parameters
    ----------
    from_member: str
        the subject member
    to_member: str
        the object member
    social_graph: dict
        the relationship data

    Returns
    -------
    str
        "follower" if from_member follows to_member,
        "followed by" if from_member is followed by to_member,
        "friends" if from_member and to_member follow each other,
        "no relationship" if neither from_member nor to_member follow each other.
    '''
    from_follows = social_graph.get(from_member, {}).get("following", [])
    to_follows = social_graph.get(to_member, {}).get("following", [])

    follow_1 = to_member in from_follows
    follow_2 = from_member in to_follows

    if follow_1 and follow_2:
        return "friends"
    if follow_1:
        return "follower"
    if follow_2:
        return "followed by"
    return "no relationship"


def tic_tac_toe(board):
    '''Tic Tac Toe.

    Tic Tac Toe is a common paper-and-pencil game.
    Players must attempt to successfully draw a straight line of their symbol across a grid.
    The player that does this first is considered the winner.

    This function evaluates a tic tac toe board and returns the winner.

    Please see "set_3_sample_data.py" for sample data. The board will adhere
    to the same pattern. The board may by 3x3, 4x4, 5x5, or 6x6. The board will never
    have more than one winner. The board will only ever have 2 unique symbols at the same time.

    Parameters
    ----------
    board: list
        the representation of the tic-tac-toe board as a square list of lists

    Returns
    -------
    str
        the symbol of the winner or "NO WINNER" if there is no winner
    '''
    n = len(board)  
    
    for row in board:
        if all(cell == row[0] and cell != '' for cell in row):
            return row[0]
    
    for col in range(n):
        first_cell = board[0][col]
        if first_cell != '' and all(board[row][col] == first_cell for row in range(n)):
            return first_cell
    
    first_cell = board[0][0]
    if first_cell != '' and all(board[i][i] == first_cell for i in range(n)):
        return first_cell
    
    first_cell = board[0][n-1]
    if first_cell != '' and all(board[i][n-1-i] == first_cell for i in range(n)):
        return first_cell
    
    return "NO WINNER"


def eta(first_stop, second_stop, route_map):
    '''ETA.

    A shuttle van service is tasked to travel along a predefined circlar route.
    This route is divided into several legs between stops.
    The route is one-way only, and it is fully connected to itself.

    This function returns how long it will take the shuttle to arrive at a stop
    after leaving another stop.

    Please see "set_3_sample_data.py" for sample data. The route map will
    adhere to the same pattern. The route map may contain more legs and more stops,
    but it will always be one-way and fully enclosed.

    Parameters
    ----------
    first_stop: str
        the stop that the shuttle will leave
    second_stop: str
        the stop that the shuttle will arrive at
    route_map: dict
        the data describing the routes

    Returns
    -------
    int
        the time it will take the shuttle to travel from first_stop to second_stop
    '''
    if first_stop == second_stop:
        return 0
    
    next_stop_map = {}
    for (from_stop, to_stop), info in route_map.items():
        next_stop_map[from_stop] = (to_stop, info['travel_time_mins'])
    
    current_stop = first_stop
    total_time = 0
    visited = set()
    
    while current_stop != second_stop:
        if current_stop in visited:
            raise ValueError(f"Invalid route: circular reference detected at {current_stop}")
        visited.add(current_stop)
        
        if current_stop not in next_stop_map:
            raise ValueError(f"No route from {current_stop}")
        
        next_stop, travel_time = next_stop_map[current_stop]
        total_time += travel_time
        current_stop = next_stop
    
    return total_time