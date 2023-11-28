import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Router } from '@angular/router';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root',
})
export class AccommodationService {
  constructor(private http: HttpClient, private router: Router) {}

  private headers = new HttpHeaders({ 'Content-Type': 'application/json' });

  getAll(): Observable<any> {
    return this.http.get('/api/accommodations/', {
      headers: this.headers,
      responseType: 'json',
    });
  }

  insert(accommodation: any): Observable<any> {
    return this.http.post(
      '/api/accommodations/create',
      {
        name: accommodation.name,
        minGuests: accommodation.minGuests,
        maxGuests: accommodation.maxGuests,
        price: accommodation.price,
      },
      { headers: this.headers, responseType: 'json' }
    );
  }
}
