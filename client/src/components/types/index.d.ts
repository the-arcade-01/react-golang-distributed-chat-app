export interface AuthResponse {
  message: string;
  data: {
    user_id: number;
    username: string;
    token: string;
  };
}

export interface Room {
  room_id: string;
  room_name: string;
  active_users: number;
}
