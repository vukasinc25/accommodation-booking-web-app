import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Router } from '@angular/router';
import { Observable } from 'rxjs';
import { Location } from '../model/location';
import { Accommodation } from '../model/accommodation';

@Injectable({
  providedIn: 'root',
})
export class AccommodationService {
  constructor(private http: HttpClient, private router: Router) {}

  private headers = new HttpHeaders({ 'Content-Type': 'application/json' });

  getById(id: number): Observable<any> {
    return this.http.get('/api/accommodations/' + id, {
      headers: this.headers,
      responseType: 'json',
    });
  }

  getAll(): Observable<any> {
    return this.http.get('/api/accommodations/', {
      headers: this.headers,
      responseType: 'json',
    });
  }

  getAllByUsername(id: string): Observable<any> {
    return this.http.get('/api/accommodations/myAccommodations/' + id, {
      headers: this.headers,
      responseType: 'json',
    });
  }

  getAllByLocation(location: string): Observable<any>{
    return this.http.get('/api/accommodations/search_by_location/' + location, {
      headers: this.headers,
      responseType: 'json',
    })
  }

  getAllByNoGuests(noGuests: string): Observable<any>{
    return this.http.get('/api/accommodations/search_by_noGuests/' + noGuests, {
      headers: this.headers,
      responseType: 'json',
    })
  }

  insert(accommodation: Accommodation): Observable<any> {
    return this.http.post(
      '/api/accommodations/create',
      {
        name: accommodation.name,
        location: {
          country: accommodation.location!.country,
          city: accommodation.location!.city,
          streetName: accommodation.location!.streetName,
          streetNumber: accommodation.location!.streetNumber,
        },
        amenities: accommodation.amenities,
        minGuests: accommodation.minGuests,
        maxGuests: accommodation.maxGuests,
        username: accommodation.username,
        // price: accommodation.price,
      },
      { headers: this.headers, responseType: 'json' }
    );
  }
}
