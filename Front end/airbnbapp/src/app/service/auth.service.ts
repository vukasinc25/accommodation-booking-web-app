import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Router } from '@angular/router';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root',
})
export class AuthService {
  constructor(private http: HttpClient, private router: Router) {}

  private headers = new HttpHeaders({ 'Content-Type': 'application/json' });

  login(user: any): Observable<any> {
    return this.http.post(
      '/api/users/login',
      { username: user.username, password: user.password },
      { headers: this.headers, responseType: 'json' }
    );
  }

  register(user: any): Observable<any> {
    return this.http.post(
      '/api/users/register',
      { username: user.username, password: user.password, role: user.userRole },
      { headers: this.headers, responseType: 'json' }
    );
  }

  logout() {
    localStorage.removeItem('jwt');
  }
}
